package edge

import (
	"database/sql"
	"fmt"
	"strconv"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Get gets edge details
func Get(projectPath, fromID, toID string) types.Result {
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

	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var e Edge
	var fromTitle, toTitle, fromStatus, toStatus string
	err = localDB.QueryRow(`
		SELECT e.from_task_id, e.to_task_id, e.created_at,
		       t1.title, t1.status, t2.title, t2.status
		FROM task_edges e
		JOIN tasks t1 ON e.from_task_id = t1.id
		JOIN tasks t2 ON e.to_task_id = t2.id
		WHERE e.from_task_id = ? AND e.to_task_id = ?
	`, fromTaskID, toTaskID).Scan(&e.FromTaskID, &e.ToTaskID, &e.CreatedAt, &fromTitle, &fromStatus, &toTitle, &toStatus)

	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("의존성을 찾을 수 없습니다: #%d -> #%d", fromTaskID, toTaskID),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}

	msg := fmt.Sprintf("의존성: #%d → #%d\n", e.FromTaskID, e.ToTaskID)
	msg += fmt.Sprintf("From: [#%d:task get %d] %s (%s)\n", e.FromTaskID, e.FromTaskID, fromTitle, fromStatus)
	msg += fmt.Sprintf("To: [#%d:task get %d] %s (%s)\n", e.ToTaskID, e.ToTaskID, toTitle, toStatus)
	msg += fmt.Sprintf("Created: %s\n", e.CreatedAt)
	msg += fmt.Sprintf("[삭제:edge delete %d %d]", e.FromTaskID, e.ToTaskID)

	return types.Result{
		Success: true,
		Message: msg,
		Data:    &e,
	}
}
