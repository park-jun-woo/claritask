package task

import (
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Add adds a new task with optional parent
func Add(projectPath, title string, parentID *int) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	// Validate parent exists if specified
	if parentID != nil {
		var exists int
		err := localDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = ?", *parentID).Scan(&exists)
		if err != nil || exists == 0 {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("부모 작업을 찾을 수 없습니다: #%d", *parentID),
			}
		}
	}

	now := db.TimeNow()
	result, err := localDB.Exec(`
		INSERT INTO tasks (parent_id, title, status, created_at)
		VALUES (?, ?, 'pending', ?)
	`, parentID, title, now)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("추가 실패: %v", err),
		}
	}

	id, _ := result.LastInsertId()

	msg := fmt.Sprintf("작업 추가됨: #%d %s", id, title)
	if parentID != nil {
		msg += fmt.Sprintf(" (부모: #%d)", *parentID)
	}
	msg += fmt.Sprintf("\n[조회:task get %d][삭제:task delete %d]", id, id)

	return types.Result{
		Success: true,
		Message: msg,
		Data: &Task{
			ID:        int(id),
			ParentID:  parentID,
			Title:     title,
			Status:    "pending",
			CreatedAt: now,
		},
	}
}
