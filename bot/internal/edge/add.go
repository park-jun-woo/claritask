package edge

import (
	"database/sql"
	"fmt"
	"strconv"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Add adds a dependency edge (fromID -> toID means toID depends on fromID)
func Add(projectPath, fromID, toID string) types.Result {
	fromTaskID, err := strconv.Atoi(fromID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("잘못된 from_task_id: %s", fromID),
		}
	}

	toTaskID, err := strconv.Atoi(toID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("잘못된 to_task_id: %s", toID),
		}
	}

	if fromTaskID == toTaskID {
		return types.Result{
			Success: false,
			Message: "자기 자신에 대한 의존성은 추가할 수 없습니다",
		}
	}

	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	// Validate from_task_id exists
	var fromTitle string
	err = localDB.QueryRow("SELECT title FROM tasks WHERE id = ?", fromTaskID).Scan(&fromTitle)
	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("from_task를 찾을 수 없습니다: #%d", fromTaskID),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("from_task 조회 실패: %v", err),
		}
	}

	// Validate to_task_id exists
	var toTitle string
	err = localDB.QueryRow("SELECT title FROM tasks WHERE id = ?", toTaskID).Scan(&toTitle)
	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("to_task를 찾을 수 없습니다: #%d", toTaskID),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("to_task 조회 실패: %v", err),
		}
	}

	// Check for existing edge
	var exists int
	err = localDB.QueryRow("SELECT 1 FROM task_edges WHERE from_task_id = ? AND to_task_id = ?", fromTaskID, toTaskID).Scan(&exists)
	if err == nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("이미 존재하는 의존성: #%d -> #%d", fromTaskID, toTaskID),
		}
	}
	if err != sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("중복 체크 실패: %v", err),
		}
	}

	// Check for cycle: BFS from toTaskID to see if fromTaskID is reachable
	cyclePath, err := findCyclePath(localDB, fromTaskID, toTaskID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("순환 검사 실패: %v", err),
		}
	}
	if cyclePath != "" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("순환 의존성 감지: %s", cyclePath),
		}
	}

	// Insert edge
	now := db.TimeNow()
	_, err = localDB.Exec(`
		INSERT INTO task_edges (from_task_id, to_task_id, created_at)
		VALUES (?, ?, ?)
	`, fromTaskID, toTaskID, now)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("의존성 추가 실패: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("의존성 추가됨: #%d(%s) → #%d(%s)\n[목록:edge list][삭제:edge delete %d %d]", fromTaskID, fromTitle, toTaskID, toTitle, fromTaskID, toTaskID),
		Data: &Edge{
			FromTaskID: fromTaskID,
			ToTaskID:   toTaskID,
			CreatedAt:  now,
		},
	}
}

// findCyclePath checks if adding edge fromTaskID→toTaskID would create a cycle.
// It performs BFS from toTaskID following existing edges to see if fromTaskID is reachable.
// Returns the cycle path string (e.g., "#3 → #2 → #1 → #3") if a cycle exists, empty string otherwise.
func findCyclePath(localDB *db.DB, fromTaskID, toTaskID int) (string, error) {
	// BFS: start from toTaskID, follow outgoing edges, see if we reach fromTaskID
	queue := []int{toTaskID}
	visited := map[int]bool{toTaskID: true}
	parent := map[int]int{toTaskID: -1}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		rows, err := localDB.Query("SELECT to_task_id FROM task_edges WHERE from_task_id = ?", current)
		if err != nil {
			return "", fmt.Errorf("간선 조회 실패: %w", err)
		}

		var neighbors []int
		for rows.Next() {
			var neighbor int
			if err := rows.Scan(&neighbor); err != nil {
				rows.Close()
				return "", fmt.Errorf("스캔 실패: %w", err)
			}
			neighbors = append(neighbors, neighbor)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return "", fmt.Errorf("행 순회 오류: %w", err)
		}

		for _, neighbor := range neighbors {
			if neighbor == fromTaskID {
				// Cycle found — reconstruct path
				parent[neighbor] = current
				return buildCyclePath(parent, fromTaskID, toTaskID), nil
			}
			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	return "", nil
}

// buildCyclePath reconstructs the cycle path from BFS parent map.
// Returns a string like "#1 → #2 → #3 → #1"
func buildCyclePath(parent map[int]int, fromTaskID, toTaskID int) string {
	// Trace back from fromTaskID to toTaskID via parent map
	// path will be [fromTaskID, ..., toTaskID]
	path := []int{fromTaskID}
	current := fromTaskID
	for current != toTaskID {
		current = parent[current]
		path = append(path, current)
	}

	// Reverse to get [toTaskID, ..., fromTaskID]
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	// Display: #fromTaskID → #toTaskID → ... → #fromTaskID
	// path = [toTaskID, ..., fromTaskID], so prepend fromTaskID to show the proposed edge
	result := fmt.Sprintf("#%d", fromTaskID)
	for _, id := range path {
		result += fmt.Sprintf(" → #%d", id)
	}
	return result
}
