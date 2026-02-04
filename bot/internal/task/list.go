package task

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/pagination"
)

// List lists tasks with pagination. If parentID is nil, shows top-level tasks (parent_id IS NULL).
// If parentID is specified, shows children of that parent.
func List(projectPath string, parentID *int, req pagination.PageRequest) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	// Count total
	var total int
	var countErr error
	if parentID == nil {
		countErr = localDB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE parent_id IS NULL`).Scan(&total)
	} else {
		countErr = localDB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE parent_id = ?`, *parentID).Scan(&total)
	}
	if countErr != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("카운트 실패: %v", countErr),
		}
	}

	var header string
	var listCmd string
	if parentID == nil {
		header = "작업"
		listCmd = "task list"
	} else {
		header = fmt.Sprintf("작업 #%d의 하위 작업", *parentID)
		listCmd = fmt.Sprintf("task list %d", *parentID)
	}

	if total == 0 {
		msg := fmt.Sprintf("%s이 없습니다.\n[추가:task add]", header)
		return types.Result{
			Success: true,
			Message: msg,
		}
	}

	var query string
	var args []interface{}
	if parentID == nil {
		query = `
			SELECT id, title, status, created_at
			FROM tasks
			WHERE parent_id IS NULL
			ORDER BY id DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{req.Limit(), req.Offset()}
	} else {
		query = `
			SELECT id, title, status, created_at
			FROM tasks
			WHERE parent_id = ?
			ORDER BY id DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{*parentID, req.Limit(), req.Offset()}
	}

	rows, queryErr := localDB.Query(query, args...)
	if queryErr != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", queryErr),
		}
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.CreatedAt); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("스캔 실패: %v", err),
			}
		}
		tasks = append(tasks, t)
	}

	pageResp := pagination.NewPageResponse(tasks, req.Page, req.PageSize, total)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 %s (%d/%d 페이지, 총 %d개)\n", header, pageResp.Page, pageResp.TotalPages, total))
	for _, t := range tasks {
		statusIcon := statusToIcon(t.Status)
		sb.WriteString(fmt.Sprintf("  %s [#%d:task get %d] %s\n", statusIcon, t.ID, t.ID, t.Title))
	}

	// Add pagination buttons
	if pageResp.HasPrev || pageResp.HasNext {
		sb.WriteString("\n")
		if pageResp.HasPrev {
			sb.WriteString(fmt.Sprintf("[◀ 이전:%s -p %d]", listCmd, pageResp.Page-1))
		}
		if pageResp.HasNext {
			sb.WriteString(fmt.Sprintf("[다음 ▶:%s -p %d]", listCmd, pageResp.Page+1))
		}
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    pageResp,
	}
}

func statusToIcon(status string) string {
	switch status {
	case "spec_ready":
		return "📝"
	case "plan_ready":
		return "📋"
	case "done":
		return "✅"
	case "failed":
		return "❌"
	default:
		return "•"
	}
}
