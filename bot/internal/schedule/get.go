package schedule

import (
	"database/sql"
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Get gets schedule details
func Get(id string) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	var s Schedule
	var enabled, runOnce int
	err = globalDB.QueryRow(`
		SELECT id, project_id, cron_expr, message, enabled, run_once, last_run, next_run, created_at, updated_at
		FROM schedules WHERE id = ?
	`, id).Scan(&s.ID, &s.ProjectID, &s.CronExpr, &s.Message, &enabled, &runOnce, &s.LastRun, &s.NextRun, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("스케줄을 찾을 수 없습니다: #%s", id),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}

	s.Enabled = enabled == 1
	s.RunOnce = runOnce == 1

	statusIcon := "✅"
	if !s.Enabled {
		statusIcon = "⏸️"
	}

	msg := fmt.Sprintf("%s 스케줄 #%d\nCron: %s\n메시지: %s\n상태: %s",
		statusIcon, s.ID, s.CronExpr, s.Message, enabledStr(s.Enabled))
	if s.RunOnce {
		msg += "\n모드: 1회 실행"
	}

	if s.ProjectID != nil {
		msg += fmt.Sprintf("\n프로젝트: %s", *s.ProjectID)
	}
	if s.LastRun != nil {
		msg += fmt.Sprintf("\n마지막 실행: %s", *s.LastRun)
	}
	if s.NextRun != nil {
		msg += fmt.Sprintf("\n다음 실행: %s", *s.NextRun)
	}

	// Action buttons
	if s.Enabled {
		msg += fmt.Sprintf("\n[비활성화:schedule disable %d][삭제:schedule delete %d]", s.ID, s.ID)
	} else {
		msg += fmt.Sprintf("\n[활성화:schedule enable %d][삭제:schedule delete %d]", s.ID, s.ID)
	}

	return types.Result{
		Success: true,
		Message: msg,
		Data:    &s,
	}
}

func enabledStr(enabled bool) string {
	if enabled {
		return "활성"
	}
	return "비활성"
}
