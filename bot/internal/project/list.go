package project

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/pagination"
)

// ListAll returns all projects without pagination
func ListAll() ([]Project, error) {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return nil, fmt.Errorf("failed to open global db: %w", err)
	}
	defer globalDB.Close()

	rows, err := globalDB.Query(`
		SELECT id, name, path, description, status, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("행 순회 오류: %w", err)
	}

	return projects, nil
}

// List lists projects with pagination
func List(req pagination.PageRequest) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to open global db: %v", err),
		}
	}
	defer globalDB.Close()

	// Count total
	var total int
	if err := globalDB.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&total); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to count projects: %v", err),
		}
	}

	if total == 0 {
		return types.Result{
			Success: true,
			Message: "프로젝트가 없습니다.\n  [생성:project create]",
		}
	}

	rows, err := globalDB.Query(`
		SELECT id, name, path, description, status, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, req.Limit(), req.Offset())
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
		if err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("failed to scan project: %v", err),
			}
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("행 순회 오류: %v", err),
		}
	}

	pageResp := pagination.NewPageResponse(projects, req.Page, req.PageSize, total)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 프로젝트 (%d/%d 페이지, 총 %d개)\n", pageResp.Page, pageResp.TotalPages, total))
	for _, p := range projects {
		sb.WriteString(fmt.Sprintf("  [%s:project switch %s]\n", p.ID, p.ID))
	}
	sb.WriteString("  [선택안함:project switch none]")

	// Add pagination buttons
	if pageResp.HasPrev || pageResp.HasNext {
		sb.WriteString("\n")
		if pageResp.HasPrev {
			sb.WriteString(fmt.Sprintf("[◀ 이전:project list -p %d]", pageResp.Page-1))
		}
		if pageResp.HasNext {
			sb.WriteString(fmt.Sprintf("[다음 ▶:project list -p %d]", pageResp.Page+1))
		}
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    pageResp,
	}
}
