package task

import (
	"database/sql"
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Run runs a task (or next pending task if id is empty)
func Run(projectPath, id string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var t Task

	if id == "" {
		// Get next pending task
		err = localDB.QueryRow(`
			SELECT id, title, content, status FROM tasks
			WHERE status = 'pending' AND parent_id IS NULL
			ORDER BY id ASC LIMIT 1
		`).Scan(&t.ID, &t.Title, &t.Content, &t.Status)
		if err == sql.ErrNoRows {
			return types.Result{
				Success: true,
				Message: "실행할 작업이 없습니다.\n[추가:task add]",
			}
		}
	} else {
		err = localDB.QueryRow(`
			SELECT id, title, content, status FROM tasks WHERE id = ?
		`, id).Scan(&t.ID, &t.Title, &t.Content, &t.Status)
		if err == sql.ErrNoRows {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("작업을 찾을 수 없습니다: #%s", id),
			}
		}
	}

	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}

	if t.Status != "pending" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 #%d은(는) 이미 %s 상태입니다.", t.ID, t.Status),
		}
	}

	// Mark as running
	now := db.TimeNow()
	_, err = localDB.Exec(`UPDATE tasks SET status = 'running', started_at = ? WHERE id = ?`, now, t.ID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("상태 업데이트 실패: %v", err),
		}
	}

	// TODO: Claude Code 실행 연동
	return types.Result{
		Success: true,
		Message: fmt.Sprintf("🔄 작업 실행 시작: #%d %s\n(Claude 연동 미구현)", t.ID, t.Title),
		Data:    &t,
	}
}
