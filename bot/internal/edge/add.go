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

	// Check for cycle (simple: reverse edge check)
	err = localDB.QueryRow("SELECT 1 FROM task_edges WHERE from_task_id = ? AND to_task_id = ?", toTaskID, fromTaskID).Scan(&exists)
	if err == nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("순환 의존성: #%d -> #%d가 이미 존재합니다", toTaskID, fromTaskID),
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
