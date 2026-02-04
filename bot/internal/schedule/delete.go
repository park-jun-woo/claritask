package schedule

import (
	"database/sql"
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Delete deletes a schedule
func Delete(id string, confirmed bool) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	// Get schedule info first
	var s Schedule
	err = globalDB.QueryRow(`SELECT id, cron_expr, message FROM schedules WHERE id = ?`, id).Scan(&s.ID, &s.CronExpr, &s.Message)
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

	// Ask for confirmation
	if !confirmed {
		return types.Result{
			Success: true,
			Message: fmt.Sprintf("스케줄 #%d '%s'을(를) 삭제하시겠습니까?\n[예:schedule delete %s yes][아니오:schedule delete %s no]",
				s.ID, truncate(s.Message, 30), id, id),
		}
	}

	// Unregister from scheduler
	if globalScheduler != nil {
		globalScheduler.Unregister(s.ID)
	}

	// Delete
	_, err = globalDB.Exec("DELETE FROM schedules WHERE id = ?", id)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("삭제 실패: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("스케줄 삭제됨: #%s", id),
	}
}
