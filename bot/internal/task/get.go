package task

import (
	"database/sql"
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Get gets task details
func Get(projectPath, id string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var t Task
	err = localDB.QueryRow(`
		SELECT id, parent_id, source, title, content, status, result, error, created_at, started_at, completed_at
		FROM tasks WHERE id = ?
	`, id).Scan(&t.ID, &t.ParentID, &t.Source, &t.Title, &t.Content, &t.Status, &t.Result, &t.Error, &t.CreatedAt, &t.StartedAt, &t.CompletedAt)

	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업을 찾을 수 없습니다: #%s", id),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}

	statusIcon := statusToIcon(t.Status)
	msg := fmt.Sprintf("%s #%d %s\nStatus: %s\nCreated: %s", statusIcon, t.ID, t.Title, t.Status, t.CreatedAt)
	if t.Content != "" {
		msg += fmt.Sprintf("\n\n%s", t.Content)
	}
	if t.Result != "" {
		msg += fmt.Sprintf("\n\nResult: %s", t.Result)
	}
	if t.Error != "" {
		msg += fmt.Sprintf("\n\nError: %s", t.Error)
	}

	// Add action buttons based on status
	switch t.Status {
	case "pending":
		msg += fmt.Sprintf("\n[실행:task run %d][삭제:task delete %d]", t.ID, t.ID)
	case "done", "failed":
		msg += fmt.Sprintf("\n[삭제:task delete %d]", t.ID)
	}

	return types.Result{
		Success: true,
		Message: msg,
		Data:    &t,
	}
}
