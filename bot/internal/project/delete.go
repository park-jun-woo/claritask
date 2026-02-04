package project

import (
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Delete deletes a project
func Delete(id string, confirmed bool) types.Result {
	// Get project info first
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
		SELECT id, name, path, type, description, status, created_at, updated_at
		FROM projects WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Path, &p.Type, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("프로젝트를 찾을 수 없습니다: %s", id),
		}
	}

	// Ask for confirmation if not confirmed
	if !confirmed {
		return types.Result{
			Success:    true,
			Message:    fmt.Sprintf("프로젝트 '%s'을(를) 삭제하시겠습니까?\nPath: %s\n[예:project delete %s yes][아니오:project delete %s no]", id, p.Path, id, id),
			NeedsInput: false,
		}
	}

	// Delete from DB
	_, err = globalDB.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to delete from db: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("프로젝트 삭제됨: %s\nPath: %s", id, p.Path),
	}
}
