package schedule

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/pagination"
)

// List lists schedules with pagination
func List(projectID *string, showAll bool, req pagination.PageRequest) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	// Count total
	var total int
	var countErr error
	if showAll {
		countErr = globalDB.QueryRow(`SELECT COUNT(*) FROM schedules`).Scan(&total)
	} else if projectID != nil {
		countErr = globalDB.QueryRow(`SELECT COUNT(*) FROM schedules WHERE project_id = ?`, *projectID).Scan(&total)
	} else {
		countErr = globalDB.QueryRow(`SELECT COUNT(*) FROM schedules WHERE project_id IS NULL`).Scan(&total)
	}
	if countErr != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("카운트 실패: %v", countErr),
		}
	}

	var header string
	var listCmd string
	if showAll {
		header = "전체 스케줄"
		listCmd = "schedule list --all"
	} else if projectID != nil {
		header = fmt.Sprintf("프로젝트 %s 스케줄", *projectID)
		listCmd = "schedule list"
	} else {
		header = "전역 스케줄"
		listCmd = "schedule list"
	}

	if total == 0 {
		msg := fmt.Sprintf("%s이 없습니다.\n[추가:schedule add]", header)
		return types.Result{
			Success: true,
			Message: msg,
		}
	}

	var query string
	var args []interface{}
	if showAll {
		query = `
			SELECT id, project_id, cron_expr, message, type, enabled, run_once, next_run
			FROM schedules
			ORDER BY id DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{req.Limit(), req.Offset()}
	} else if projectID != nil {
		query = `
			SELECT id, project_id, cron_expr, message, type, enabled, run_once, next_run
			FROM schedules
			WHERE project_id = ?
			ORDER BY id DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{*projectID, req.Limit(), req.Offset()}
	} else {
		query = `
			SELECT id, project_id, cron_expr, message, type, enabled, run_once, next_run
			FROM schedules
			WHERE project_id IS NULL
			ORDER BY id DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{req.Limit(), req.Offset()}
	}

	rows, queryErr := globalDB.Query(query, args...)
	if queryErr != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", queryErr),
		}
	}
	defer rows.Close()

	var schedules []Schedule
	for rows.Next() {
		var s Schedule
		var enabled, runOnce int
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.CronExpr, &s.Message, &s.Type, &enabled, &runOnce, &s.NextRun); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("스캔 실패: %v", err),
			}
		}
		s.Enabled = enabled == 1
		s.RunOnce = runOnce == 1
		schedules = append(schedules, s)
	}
	if err := rows.Err(); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("행 순회 오류: %v", err),
		}
	}

	pageResp := pagination.NewPageResponse(schedules, req.Page, req.PageSize, total)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📅 %s (%d/%d 페이지, 총 %d개)\n", header, pageResp.Page, pageResp.TotalPages, total))
	for _, s := range schedules {
		statusIcon := "✅"
		if !s.Enabled {
			statusIcon = "⏸️"
		}
		onceMarker := ""
		if s.RunOnce {
			onceMarker = " [1회]"
		}
		typeMarker := ""
		if s.Type == "bash" {
			typeMarker = " [bash]"
		}
		sb.WriteString(fmt.Sprintf("  %s [#%d:schedule get %d] %s %s%s%s\n",
			statusIcon, s.ID, s.ID, s.CronExpr, truncate(s.Message, 30), typeMarker, onceMarker))
	}

	// Pagination buttons
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
