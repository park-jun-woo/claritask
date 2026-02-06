package project

import (
	"database/sql"
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Switch switches to a project
func Switch(id string) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to open global db: %v", err),
		}
	}
	defer globalDB.Close()

	var p Project
	err = globalDB.QueryRow(`
		SELECT id, name, path, description, status, created_at, updated_at
		FROM projects WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Path, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("프로젝트를 찾을 수 없습니다: %s", id),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to query project: %v", err),
		}
	}

	msg := fmt.Sprintf("프로젝트 선택됨: %s\nPath: %s\n  [삭제:project delete %s]", p.ID, p.Path, p.ID)

	return types.Result{
		Success: true,
		Message: msg,
		Data:    &p,
	}
}
