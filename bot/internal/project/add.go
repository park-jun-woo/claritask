package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Add adds an existing folder as a project
func Add(path, projType, description string) types.Result {
	// Require path
	if path == "" {
		return types.Result{
			Success:    true,
			Message:    "프로젝트 경로를 입력하세요:",
			NeedsInput: true,
			Prompt:     "Path: ",
			Context:    "project add",
		}
	}

	// Expand ~ to home directory
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("경로 변환 실패: %v", err),
		}
	}

	// Check if folder exists
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("폴더가 존재하지 않습니다: %s", absPath),
		}
	}
	if !info.IsDir() {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("디렉토리가 아닙니다: %s", absPath),
		}
	}

	// Require type - ask for input if missing
	if projType == "" {
		return types.Result{
			Success:    true,
			Message:    fmt.Sprintf("Select project type:\n%s", FormatTypeList()),
			NeedsInput: true,
			Prompt:     "Type: ",
			Context:    "project add " + absPath,
		}
	}

	// Validate type
	validType := false
	typeIDs := GetTypeIDs()
	for _, t := range typeIDs {
		if t == projType {
			validType = true
			break
		}
	}
	if !validType {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("invalid type: %s\nvalid types: %s", projType, strings.Join(typeIDs, ", ")),
		}
	}

	// Require description - ask for input if missing
	if description == "" {
		return types.Result{
			Success:    true,
			Message:    "Enter project description:",
			NeedsInput: true,
			Prompt:     "Description: ",
			Context:    "project add " + absPath + " " + projType,
		}
	}

	// Use folder name as project ID
	id := filepath.Base(absPath)

	// Check if already registered in global DB
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to open global db: %v", err),
		}
	}
	defer globalDB.Close()

	// Check for duplicate ID or path
	var existingID string
	err = globalDB.QueryRow(`SELECT id FROM projects WHERE id = ? OR path = ?`, id, absPath).Scan(&existingID)
	if err == nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("이미 등록된 프로젝트입니다: %s", existingID),
		}
	}

	// Register in global DB
	now := db.TimeNow()
	_, err = globalDB.Exec(`
		INSERT INTO projects (id, name, path, type, description, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'active', ?, ?)
	`, id, id, absPath, projType, description, now, now)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to insert project: %v", err),
		}
	}

	// Create local DB and migrate
	localDB, err := db.OpenLocal(absPath)
	if err != nil {
		// Rollback global DB entry
		globalDB.Exec(`DELETE FROM projects WHERE id = ?`, id)
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to create local db: %v", err),
		}
	}
	defer localDB.Close()

	if err := localDB.MigrateLocal(); err != nil {
		// Rollback global DB entry
		globalDB.Exec(`DELETE FROM projects WHERE id = ?`, id)
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to migrate local db: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("프로젝트 추가됨: %s\nPath: %s\nType: %s\n%s\n[전환:project switch %s][삭제:project delete %s]", id, absPath, projType, description, id, id),
		Data: &Project{
			ID:          id,
			Name:        id,
			Path:        absPath,
			Type:        projType,
			Description: description,
		},
	}
}
