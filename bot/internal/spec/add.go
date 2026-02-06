package spec

import (
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Add adds a new spec
func Add(projectPath, title, content string) types.Result {
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
		INSERT INTO specs (title, content, status, created_at, updated_at)
		VALUES (?, ?, 'draft', ?, ?)
	`, title, content, now, now)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("추가 실패: %v", err),
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("ID 획득 실패: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("스펙 추가됨: #%d %s", id, title),
		Data: &Spec{
			ID:        int(id),
			Title:     title,
			Content:   content,
			Status:    "draft",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}
