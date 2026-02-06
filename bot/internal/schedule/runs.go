package schedule

import (
	"database/sql"
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/pagination"
)

// Runs lists schedule runs for a schedule with pagination
func Runs(scheduleID string, req pagination.PageRequest) types.Result {
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
	err = globalDB.QueryRow(`SELECT COUNT(*) FROM schedule_runs WHERE schedule_id = ?`, scheduleID).Scan(&total)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("카운트 실패: %v", err),
		}
	}

	if total == 0 {
		return types.Result{
			Success: true,
			Message: fmt.Sprintf("스케줄 #%s 실행 기록이 없습니다.", scheduleID),
		}
	}

	rows, err := globalDB.Query(`
		SELECT id, schedule_id, status, result, error, started_at, completed_at
		FROM schedule_runs
		WHERE schedule_id = ?
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, scheduleID, req.Limit(), req.Offset())
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}
	defer rows.Close()

	var runs []ScheduleRun
	for rows.Next() {
		var r ScheduleRun
		if err := rows.Scan(&r.ID, &r.ScheduleID, &r.Status, &r.Result, &r.Error, &r.StartedAt, &r.CompletedAt); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("스캔 실패: %v", err),
			}
		}
		runs = append(runs, r)
	}
	if err := rows.Err(); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("행 순회 오류: %v", err),
		}
	}

	pageResp := pagination.NewPageResponse(runs, req.Page, req.PageSize, total)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 스케줄 #%s 실행 기록 (%d/%d 페이지, 총 %d개)\n", scheduleID, pageResp.Page, pageResp.TotalPages, total))
	for _, r := range runs {
		statusIcon := statusToIcon(r.Status)
		sb.WriteString(fmt.Sprintf("  %s [#%d:schedule run %d] %s %s\n",
			statusIcon, r.ID, r.ID, r.Status, r.StartedAt))
	}

	// Pagination buttons
	if pageResp.HasPrev || pageResp.HasNext {
		sb.WriteString("\n")
		if pageResp.HasPrev {
			sb.WriteString(fmt.Sprintf("[◀ 이전:schedule runs %s -p %d]", scheduleID, pageResp.Page-1))
		}
		if pageResp.HasNext {
			sb.WriteString(fmt.Sprintf("[다음 ▶:schedule runs %s -p %d]", scheduleID, pageResp.Page+1))
		}
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    pageResp,
	}
}

// Run gets a single schedule run detail
func Run(runID string) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	var r ScheduleRun
	err = globalDB.QueryRow(`
		SELECT id, schedule_id, status, result, error, started_at, completed_at
		FROM schedule_runs WHERE id = ?
	`, runID).Scan(&r.ID, &r.ScheduleID, &r.Status, &r.Result, &r.Error, &r.StartedAt, &r.CompletedAt)

	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("실행 기록을 찾을 수 없습니다: #%s", runID),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}

	statusIcon := statusToIcon(r.Status)
	msg := fmt.Sprintf("%s 실행 #%d (스케줄 #%d)\n상태: %s\n시작: %s",
		statusIcon, r.ID, r.ScheduleID, r.Status, r.StartedAt)

	if r.CompletedAt != nil {
		msg += fmt.Sprintf("\n완료: %s", *r.CompletedAt)
	}

	if r.Result != "" {
		msg += fmt.Sprintf("\n\n📄 결과:\n%s", truncate(r.Result, 1000))
	}

	if r.Error != "" {
		msg += fmt.Sprintf("\n\n❌ 에러:\n%s", r.Error)
	}

	msg += fmt.Sprintf("\n\n[스케줄 보기:schedule get %d][실행 기록:schedule runs %d]", r.ScheduleID, r.ScheduleID)

	return types.Result{
		Success: true,
		Message: msg,
		Data:    &r,
	}
}

func statusToIcon(status string) string {
	switch status {
	case "running":
		return "🔄"
	case "done":
		return "✅"
	case "failed":
		return "❌"
	default:
		return "❓"
	}
}
