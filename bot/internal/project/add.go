package project

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"parkjunwoo.com/claribot/internal/config"
	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Add adds an existing folder as a project
func Add(path, description string) types.Result {
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

	// Require description - ask for input if missing
	if description == "" {
		return types.Result{
			Success:    true,
			Message:    "Enter project description:",
			NeedsInput: true,
			Prompt:     "Description: ",
			Context:    "project add " + absPath,
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
		INSERT INTO projects (id, name, path, description, status, category, pinned, last_accessed, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'active', '', 0, '', ?, ?)
	`, id, id, absPath, description, now, now)
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

	// Set default parallel from config
	cfg, _ := config.Load()
	defaultParallel := cfg.Project.DefaultParallel
	now = db.TimeNow()
	localDB.Exec(
		"INSERT OR REPLACE INTO config (key, value, updated_at) VALUES (?, ?, ?)",
		"parallel", strconv.Itoa(defaultParallel), now,
	)

	// GitHub repo auto-create (best-effort)
	if err := ensureGitHub(absPath, id); err != nil {
		log.Printf("[project] GitHub 연결 실패 (무시): %v", err)
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("프로젝트 추가됨: %s\nPath: %s\n%s\n[전환:project switch %s][삭제:project delete %s]", id, absPath, description, id, id),
		Data: &Project{
			ID:          id,
			Name:        id,
			Path:        absPath,
			Description: description,
		},
	}
}
