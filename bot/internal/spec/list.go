package spec

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/pagination"
)

// List lists specs with pagination
func List(projectPath string, req pagination.PageRequest) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var total int
	if err := localDB.QueryRow(`SELECT COUNT(*) FROM specs`).Scan(&total); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("카운트 실패: %v", err),
		}
	}

	if total == 0 {
		return types.Result{
			Success: true,
			Message: "스펙이 없습니다.\n[추가:spec add]",
		}
	}

	rows, err := localDB.Query(`
		SELECT id, title, status, priority, created_at
		FROM specs
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, req.Limit(), req.Offset())
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}
	defer rows.Close()

	var specs []Spec
	for rows.Next() {
		var s Spec
		if err := rows.Scan(&s.ID, &s.Title, &s.Status, &s.Priority, &s.CreatedAt); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("스캔 실패: %v", err),
			}
		}
		specs = append(specs, s)
	}
	if err := rows.Err(); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("행 순회 오류: %v", err),
		}
	}

	pageResp := pagination.NewPageResponse(specs, req.Page, req.PageSize, total)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 스펙 (%d/%d 페이지, 총 %d개)\n", pageResp.Page, pageResp.TotalPages, total))
	for _, s := range specs {
		icon := statusToIcon(s.Status)
		sb.WriteString(fmt.Sprintf("  %s [#%d:spec get %d] %s\n", icon, s.ID, s.ID, s.Title))
	}

	if pageResp.HasPrev || pageResp.HasNext {
		sb.WriteString("\n")
		if pageResp.HasPrev {
			sb.WriteString(fmt.Sprintf("[◀ 이전:spec list -p %d]", pageResp.Page-1))
		}
		if pageResp.HasNext {
			sb.WriteString(fmt.Sprintf("[다음 ▶:spec list -p %d]", pageResp.Page+1))
		}
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    pageResp,
	}
}
