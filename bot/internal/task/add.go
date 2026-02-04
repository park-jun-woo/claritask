package task

import (
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Add adds a new task
func Add(projectPath, title string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	now := db.TimeNow()
	result, err := localDB.Exec(`
		INSERT INTO tasks (title, status, created_at)
		VALUES (?, 'pending', ?)
	`, title, now)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("추가 실패: %v", err),
		}
	}

	id, _ := result.LastInsertId()

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("작업 추가됨: #%d %s\n[조회:task get %d][삭제:task delete %d]", id, title, id, id),
		Data: &Task{
			ID:        int(id),
			Title:     title,
			Status:    "pending",
			CreatedAt: now,
		},
	}
}
