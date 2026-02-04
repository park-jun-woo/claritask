package schedule

import (
	"database/sql"
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// SetProject updates the project of a schedule
func SetProject(id string, projectID *string) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	// Check schedule exists
	var exists int
	err = globalDB.QueryRow(`SELECT COUNT(*) FROM schedules WHERE id = ?`, id).Scan(&exists)
	if err != nil || exists == 0 {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("스케줄을 찾을 수 없습니다: #%s", id),
		}
	}

	// Validate project exists if specified
	if projectID != nil {
		var projectExists int
		err := globalDB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", *projectID).Scan(&projectExists)
		if err != nil || projectExists == 0 {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("프로젝트를 찾을 수 없습니다: %s", *projectID),
			}
		}
	}

	now := db.TimeNow()
	_, err = globalDB.Exec(`UPDATE schedules SET project_id = ?, updated_at = ? WHERE id = ?`,
		projectID, now, id)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("업데이트 실패: %v", err),
		}
	}

	// Re-register with scheduler to update the job
	var s Schedule
	var enabled, runOnce int
	err = globalDB.QueryRow(`SELECT id, project_id, cron_expr, message, enabled, run_once FROM schedules WHERE id = ?`, id).
		Scan(&s.ID, &s.ProjectID, &s.CronExpr, &s.Message, &enabled, &runOnce)
	if err != nil && err != sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("스케줄 조회 실패: %v", err),
		}
	}
	s.Enabled = enabled == 1
	s.RunOnce = runOnce == 1

	if globalScheduler != nil && s.Enabled {
		globalScheduler.Register(s.ID, s.CronExpr, s.Message, s.ProjectID, s.RunOnce)
	}

	var msg string
	if projectID == nil {
		msg = fmt.Sprintf("스케줄 #%s 프로젝트 해제됨 (글로벌)", id)
	} else {
		msg = fmt.Sprintf("스케줄 #%s 프로젝트 변경됨: %s", id, *projectID)
	}
	msg += fmt.Sprintf("\n[조회:schedule get %s]", id)

	return types.Result{
		Success: true,
		Message: msg,
	}
}
