package schedule

import (
	"fmt"
	"unicode/utf8"

	"github.com/robfig/cron/v3"
	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Add adds a new schedule
func Add(cronExpr, message string, projectID *string, runOnce bool, scheduleType string) types.Result {
	if scheduleType == "" {
		scheduleType = "claude"
	}
	if scheduleType != "claude" && scheduleType != "bash" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("잘못된 스케줄 타입: %s (claude 또는 bash)", scheduleType),
		}
	}
	// Validate cron expression
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("잘못된 cron 표현식: %v", err),
		}
	}

	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	// Validate project exists if specified
	if projectID != nil {
		var exists int
		err := globalDB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", *projectID).Scan(&exists)
		if err != nil || exists == 0 {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("프로젝트를 찾을 수 없습니다: %s", *projectID),
			}
		}
	}

	now := db.TimeNow()
	nextRun := schedule.Next(parseTime(now)).Format("2006-01-02T15:04:05Z07:00")

	runOnceInt := 0
	if runOnce {
		runOnceInt = 1
	}

	result, err := globalDB.Exec(`
		INSERT INTO schedules (project_id, cron_expr, message, type, enabled, run_once, next_run, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, ?, ?, ?, ?)
	`, projectID, cronExpr, message, scheduleType, runOnceInt, nextRun, now, now)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("추가 실패: %v", err),
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("스케줄 ID 획득 실패: %v", err),
		}
	}

	// Register with global scheduler
	if globalScheduler != nil {
		globalScheduler.Register(int(id), cronExpr, message, projectID, runOnce, scheduleType)
	}

	msg := fmt.Sprintf("스케줄 추가됨: #%d\nCron: %s\n타입: %s\n메시지: %s\n다음 실행: %s", id, cronExpr, scheduleType, truncate(message, 50), nextRun)
	if runOnce {
		msg += "\n모드: 1회 실행"
	}
	if projectID != nil {
		msg += fmt.Sprintf("\n프로젝트: %s", *projectID)
	}
	msg += fmt.Sprintf("\n[조회:schedule get %d][삭제:schedule delete %d]", id, id)

	return types.Result{
		Success: true,
		Message: msg,
		Data: &Schedule{
			ID:        int(id),
			ProjectID: projectID,
			CronExpr:  cronExpr,
			Message:   message,
			Type:      scheduleType,
			Enabled:   true,
			RunOnce:   runOnce,
			NextRun:   &nextRun,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func truncate(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	return string([]rune(s)[:maxLen]) + "..."
}
