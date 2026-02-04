package project

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// List lists all projects
func List() types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to open global db: %v", err),
		}
	}
	defer globalDB.Close()

	rows, err := globalDB.Query(`
		SELECT id, name, path, type, description, status, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to query projects: %v", err),
		}
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.Type, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("failed to scan project: %v", err),
			}
		}
		projects = append(projects, p)
	}

	if len(projects) == 0 {
		return types.Result{
			Success: true,
			Message: "프로젝트가 없습니다.\n  [생성:project create]",
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("프로젝트 (%d):\n", len(projects)))
	for _, p := range projects {
		sb.WriteString(fmt.Sprintf("  [%s:project switch %s]\n", p.ID, p.ID))
	}
	sb.WriteString("  [선택안함:project switch none]")

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    projects,
	}
}
