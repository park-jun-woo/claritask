package task

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
)

// BuildContextMap builds a text summary of the entire task tree with edge info.
// Used as lightweight context instead of injecting full related task content.
func BuildContextMap(localDB *db.DB) (string, error) {
	rows, err := localDB.Query(`
		SELECT id, parent_id, title, status, depth
		FROM tasks
		ORDER BY depth ASC, id ASC
	`)
	if err != nil {
		return "", fmt.Errorf("task 조회 실패: %w", err)
	}
	defer rows.Close()

	type taskInfo struct {
		ID       int
		ParentID *int
		Title    string
		Status   string
		Depth    int
	}

	var tasks []taskInfo
	for rows.Next() {
		var t taskInfo
		if err := rows.Scan(&t.ID, &t.ParentID, &t.Title, &t.Status, &t.Depth); err != nil {
			return "", fmt.Errorf("task 스캔 실패: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("task 행 순회 오류: %w", err)
	}

	if len(tasks) == 0 {
		return "(작업 없음)", nil
	}

	// Build edge map: from_task_id -> []to_task_id
	edgeMap := make(map[int][]int)
	edgeRows, err := localDB.Query(`SELECT from_task_id, to_task_id FROM task_edges`)
	if err != nil {
		return "", fmt.Errorf("edge 조회 실패: %w", err)
	}
	defer edgeRows.Close()

	for edgeRows.Next() {
		var from, to int
		if err := edgeRows.Scan(&from, &to); err != nil {
			return "", fmt.Errorf("edge 스캔 실패: %w", err)
		}
		edgeMap[from] = append(edgeMap[from], to)
	}
	if err := edgeRows.Err(); err != nil {
		return "", fmt.Errorf("edge 행 순회 오류: %w", err)
	}

	// Format output
	var sb strings.Builder
	for _, t := range tasks {
		indent := strings.Repeat("  ", t.Depth)
		sb.WriteString(fmt.Sprintf("%s#%d [%s] %s", indent, t.ID, t.Status, t.Title))

		if deps, ok := edgeMap[t.ID]; ok {
			depStrs := make([]string, len(deps))
			for i, d := range deps {
				depStrs[i] = fmt.Sprintf("#%d", d)
			}
			sb.WriteString(fmt.Sprintf(" → depends on: %s", strings.Join(depStrs, ", ")))
		}

		sb.WriteString("\n")
	}

	return sb.String(), nil
}
