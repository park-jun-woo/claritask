package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"parkjunwoo.com/talos/internal/db"
	"parkjunwoo.com/talos/internal/service"
)

var initCmd = &cobra.Command{
	Use:   "init <project-id> [description]",
	Short: "Initialize a new project",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	description := ""
	if len(args) > 1 {
		description = args[1]
	}

	// Validate project ID
	if err := validateProjectID(projectID); err != nil {
		outputError(err)
		return nil
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		outputError(fmt.Errorf("get working directory: %w", err))
		return nil
	}

	projectPath := filepath.Join(cwd, projectID)
	talosPath := filepath.Join(projectPath, ".talos")
	dbPath := filepath.Join(talosPath, "db")

	// Check if directory already exists
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		outputError(fmt.Errorf("directory already exists: %s", projectPath))
		return nil
	}

	// Create project directory
	if err := os.MkdirAll(talosPath, 0755); err != nil {
		outputError(fmt.Errorf("create directory: %w", err))
		return nil
	}

	// Open database and run migrations
	database, err := db.Open(dbPath)
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		outputError(fmt.Errorf("migrate database: %w", err))
		return nil
	}

	// Create project in database
	if err := service.CreateProject(database, projectID, projectID, description); err != nil {
		outputError(fmt.Errorf("create project: %w", err))
		return nil
	}

	// Initialize state
	if err := service.InitState(database, projectID); err != nil {
		outputError(fmt.Errorf("init state: %w", err))
		return nil
	}

	// Create CLAUDE.md
	claudeContent := fmt.Sprintf(claudeTemplate, projectID, description)
	claudePath := filepath.Join(projectPath, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte(claudeContent), 0644); err != nil {
		outputError(fmt.Errorf("create CLAUDE.md: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":    true,
		"project_id": projectID,
		"path":       projectPath,
		"message":    "Project initialized successfully",
	})

	return nil
}

func validateProjectID(id string) error {
	matched, _ := regexp.MatchString(`^[a-z0-9_-]+$`, id)
	if !matched {
		return fmt.Errorf("invalid project ID: %s (only lowercase letters, numbers, hyphens, and underscores allowed)", id)
	}
	return nil
}

const claudeTemplate = `# %s

## Description
%s

## Tech Stack
- Backend:
- Frontend:
- Database:

## Commands
- ` + "`talos project set '<json>'`" + ` - 프로젝트 설정
- ` + "`talos required`" + ` - 필수 입력 확인
- ` + "`talos project plan`" + ` - 플래닝 시작
- ` + "`talos project start`" + ` - 실행 시작
`
