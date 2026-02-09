package project

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"parkjunwoo.com/claribot/internal/config"
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

	// Set default parallel from config
	cfg, _ := config.Load()
	defaultParallel := cfg.Project.DefaultParallel
	now = db.TimeNow()
	localDB.Exec(
		"INSERT OR REPLACE INTO config (key, value, updated_at) VALUES (?, ?, ?)",
		"parallel", strconv.Itoa(defaultParallel), now,
	)

	// GitHub repo auto-create (best-effort)
	if err := ensureGitHub(projectPath, id); err != nil {
		log.Printf("[project] GitHub 연결 실패 (무시): %v", err)
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

// ensureGitHub initializes git and creates a private GitHub repo if not already set up.
func ensureGitHub(projectPath, projectID string) error {
	// 1. git init (skip if .git already exists)
	if _, err := os.Stat(filepath.Join(projectPath, ".git")); os.IsNotExist(err) {
		if err := exec.Command("git", "-C", projectPath, "init").Run(); err != nil {
			return fmt.Errorf("git init failed: %w", err)
		}
	}

	// 2. Check if remote origin already exists
	out, _ := exec.Command("git", "-C", projectPath, "remote", "get-url", "origin").Output()
	if len(strings.TrimSpace(string(out))) > 0 {
		return nil // remote already configured
	}

	// 3. Create private GitHub repo
	return exec.Command("gh", "repo", "create", projectID,
		"--private", "--source", projectPath, "--remote", "origin").Run()
}
