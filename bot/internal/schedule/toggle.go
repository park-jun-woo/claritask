package schedule

import (
	"database/sql"
	"fmt"

	"github.com/robfig/cron/v3"
	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Enable enables a schedule
func Enable(id string) types.Result {
	return setEnabled(id, true)
}

// Disable disables a schedule
func Disable(id string) types.Result {
	return setEnabled(id, false)
}

func setEnabled(id string, enabled bool) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	// Get schedule info
	var s Schedule
	var runOnce int
	err = globalDB.QueryRow(`SELECT id, project_id, cron_expr, message, type, run_once FROM schedules WHERE id = ?`, id).Scan(&s.ID, &s.ProjectID, &s.CronExpr, &s.Message, &s.Type, &runOnce)
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
	s.RunOnce = runOnce == 1

	now := db.TimeNow()
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}

	// Calculate next run if enabling
	var nextRun *string
	if enabled {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		schedule, _ := parser.Parse(s.CronExpr)
		next := schedule.Next(parseTime(now)).Format("2006-01-02T15:04:05Z07:00")
		nextRun = &next
	}

	_, err = globalDB.Exec(`UPDATE schedules SET enabled = ?, next_run = ?, updated_at = ? WHERE id = ?`,
		enabledInt, nextRun, now, id)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("업데이트 실패: %v", err),
		}
	}

	// Update scheduler
	if globalScheduler != nil {
		if enabled {
			globalScheduler.Register(s.ID, s.CronExpr, s.Message, s.ProjectID, s.RunOnce, s.Type)
		} else {
			globalScheduler.Unregister(s.ID)
		}
	}

	action := "활성화"
	if !enabled {
		action = "비활성화"
	}

	msg := fmt.Sprintf("스케줄 #%s %s됨\n[조회:schedule get %s]", id, action, id)
	return types.Result{
		Success: true,
		Message: msg,
	}
}
