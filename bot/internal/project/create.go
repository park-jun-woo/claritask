package project

import (
	"fmt"
	"os"
	"path/filepath"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Create creates a new project
func Create(id, description string) types.Result {
	// Check default path
	if DefaultPath == "" {
		return types.Result{
			Success: false,
			Message: "project.path not configured in config.yaml",
		}
	}

	// Require description - ask for input if missing
	if description == "" {
		return types.Result{
			Success:    true,
			Message:    "Enter project description:",
			NeedsInput: true,
			Prompt:     "Description: ",
			Context:    "project create " + id,
		}
	}

	// Build project path
	projectPath := filepath.Join(DefaultPath, id)

	// Check if already exists
	if _, err := os.Stat(projectPath); err == nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("project already exists: %s", projectPath),
		}
	}

	// 1. Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to create directory: %v", err),
		}
	}

	// 2. Add to global DB
	globalDB, err := db.OpenGlobal()
	if err != nil {
		os.RemoveAll(projectPath) // rollback
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to open global db: %v", err),
		}
	}
	defer globalDB.Close()

	now := db.TimeNow()
	_, err = globalDB.Exec(`
		INSERT INTO projects (id, name, path, description, status, category, pinned, last_accessed, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'active', '', 0, '', ?, ?)
	`, id, id, projectPath, description, now, now)
	if err != nil {
		os.RemoveAll(projectPath) // rollback
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to insert project: %v", err),
		}
	}

	// 3. Create local DB and migrate
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		globalDB.Exec(`DELETE FROM projects WHERE id = ?`, id) // rollback
		os.RemoveAll(projectPath)                              // rollback
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to create local db: %v", err),
		}
	}
	defer localDB.Close()

	if err := localDB.MigrateLocal(); err != nil {
		globalDB.Exec(`DELETE FROM projects WHERE id = ?`, id) // rollback
		os.RemoveAll(projectPath)                              // rollback
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to migrate local db: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("프로젝트 생성됨: %s\n%s\n[삭제:project delete %s]", id, description, id),
		Data: &Project{
			ID:          id,
			Name:        id,
			Path:        projectPath,
			Description: description,
		},
	}
}
