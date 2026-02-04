package schedule

import (
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/project"
	"parkjunwoo.com/claribot/internal/prompts"
	"parkjunwoo.com/claribot/pkg/claude"
)

// globalScheduler is the singleton scheduler instance
var globalScheduler *Scheduler

// Scheduler manages cron jobs for schedules
type Scheduler struct {
	cron     *cron.Cron
	jobs     map[int]cron.EntryID // schedule ID -> cron entry ID
	mu       sync.RWMutex
	notifier func(projectID *string, msg string) // 텔레그램 알림 콜백
}

// Init initializes the global scheduler
func Init(notifier func(projectID *string, msg string)) error {
	globalScheduler = &Scheduler{
		cron:     cron.New(cron.WithParser(cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow))),
		jobs:     make(map[int]cron.EntryID),
		notifier: notifier,
	}

	// Load existing schedules from DB
	if err := globalScheduler.loadFromDB(); err != nil {
		return err
	}

	globalScheduler.cron.Start()
	log.Printf("Scheduler started with %d jobs", len(globalScheduler.jobs))
	return nil
}

// Shutdown stops the scheduler
func Shutdown() {
	if globalScheduler != nil {
		globalScheduler.cron.Stop()
		log.Println("Scheduler stopped")
	}
}

// loadFromDB loads all enabled schedules from database
func (s *Scheduler) loadFromDB() error {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return err
	}
	defer globalDB.Close()

	rows, err := globalDB.Query(`
		SELECT id, project_id, cron_expr, message, run_once
		FROM schedules
		WHERE enabled = 1
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var projectID *string
		var cronExpr, msg string
		var runOnce int
		if err := rows.Scan(&id, &projectID, &cronExpr, &msg, &runOnce); err != nil {
			log.Printf("Scheduler: 스케줄 로드 실패: %v", err)
			continue
		}
		s.Register(id, cronExpr, msg, projectID, runOnce == 1)
	}

	return rows.Err()
}

// Register adds a schedule to the cron
func (s *Scheduler) Register(scheduleID int, cronExpr, msg string, projectID *string, runOnce bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing job if any
	if entryID, exists := s.jobs[scheduleID]; exists {
		s.cron.Remove(entryID)
	}

	// Add new job
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.execute(scheduleID, msg, projectID, runOnce)
	})
	if err != nil {
		log.Printf("Scheduler: 스케줄 #%d 등록 실패: %v", scheduleID, err)
		return
	}

	s.jobs[scheduleID] = entryID
	if runOnce {
		log.Printf("Scheduler: 스케줄 #%d 등록됨 (cron: %s, 1회 실행)", scheduleID, cronExpr)
	} else {
		log.Printf("Scheduler: 스케줄 #%d 등록됨 (cron: %s)", scheduleID, cronExpr)
	}
}

// Unregister removes a schedule from the cron
func (s *Scheduler) Unregister(scheduleID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.jobs[scheduleID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, scheduleID)
		log.Printf("Scheduler: 스케줄 #%d 제거됨", scheduleID)
	}
}

// execute runs a scheduled task with Claude Code
func (s *Scheduler) execute(scheduleID int, msg string, projectID *string, runOnce bool) {
	log.Printf("Scheduler: 스케줄 #%d 실행 시작", scheduleID)

	globalDB, err := db.OpenGlobal()
	if err != nil {
		log.Printf("Scheduler: DB 열기 실패: %v", err)
		return
	}
	defer globalDB.Close()

	// Create schedule_run record with 'running' status
	startedAt := db.TimeNow()
	result, err := globalDB.Exec(`
		INSERT INTO schedule_runs (schedule_id, status, started_at)
		VALUES (?, 'running', ?)
	`, scheduleID, startedAt)
	if err != nil {
		log.Printf("Scheduler: schedule_run 생성 실패: %v", err)
		return
	}
	runID, _ := result.LastInsertId()

	// Auto-disable if run_once (before Claude execution to prevent re-runs on error)
	if runOnce {
		_, err = globalDB.Exec(`UPDATE schedules SET enabled = 0, updated_at = ? WHERE id = ?`, db.TimeNow(), scheduleID)
		if err != nil {
			log.Printf("Scheduler: 스케줄 #%d 자동 비활성화 실패: %v", scheduleID, err)
		} else {
			log.Printf("Scheduler: 스케줄 #%d 1회 실행, 자동 비활성화됨", scheduleID)
			s.Unregister(scheduleID)
		}
	}

	// Get project path
	var projectPath string
	if projectID != nil {
		projResult := project.Get(*projectID)
		if projResult.Success {
			if p, ok := projResult.Data.(*project.Project); ok {
				projectPath = p.Path
			}
		}
	}
	if projectPath == "" {
		projectPath = project.DefaultPath
	}

	// Load system prompt
	systemPrompt, err := prompts.Get(prompts.Common, "schedule")
	if err != nil {
		log.Printf("Scheduler: 시스템 프롬프트 로드 실패: %v", err)
		systemPrompt = ""
	}

	// Execute Claude Code
	opts := claude.Options{
		UserPrompt:   msg,
		SystemPrompt: systemPrompt,
		WorkDir:      projectPath,
	}

	claudeResult, claudeErr := claude.Run(opts)

	completedAt := db.TimeNow()
	var status, resultText, errorText string

	if claudeErr != nil {
		status = "failed"
		errorText = claudeErr.Error()
		log.Printf("Scheduler: 스케줄 #%d Claude 실행 실패: %v", scheduleID, claudeErr)
	} else if claudeResult.ExitCode != 0 {
		status = "failed"
		resultText = claudeResult.Output
		errorText = "exit code: " + string(rune(claudeResult.ExitCode))
		log.Printf("Scheduler: 스케줄 #%d Claude 비정상 종료 (exit: %d)", scheduleID, claudeResult.ExitCode)
	} else {
		status = "done"
		resultText = claudeResult.Output
		log.Printf("Scheduler: 스케줄 #%d 실행 완료", scheduleID)
	}

	// Update schedule_run with result
	_, err = globalDB.Exec(`
		UPDATE schedule_runs
		SET status = ?, result = ?, error = ?, completed_at = ?
		WHERE id = ?
	`, status, resultText, errorText, completedAt, runID)
	if err != nil {
		log.Printf("Scheduler: schedule_run 업데이트 실패: %v", err)
	}

	// Update schedules last_run and next_run
	s.updateRunTimes(scheduleID, globalDB)

	// Send notification
	if s.notifier != nil {
		var notification string
		if status == "done" {
			notification = "📅 스케줄 실행 완료: " + truncate(msg, 50) + "\n\n" + truncate(resultText, 500)
		} else {
			notification = "❌ 스케줄 실행 실패: " + truncate(msg, 50) + "\n\n" + errorText
		}
		s.notifier(projectID, notification)
	}
}

// updateRunTimes updates last_run and next_run for a schedule
func (s *Scheduler) updateRunTimes(scheduleID int, globalDB *db.DB) {
	// Get cron expression
	var cronExpr string
	err := globalDB.QueryRow(`SELECT cron_expr FROM schedules WHERE id = ?`, scheduleID).Scan(&cronExpr)
	if err != nil {
		log.Printf("Scheduler: 스케줄 #%d 조회 실패: %v", scheduleID, err)
		return
	}

	now := db.TimeNow()

	// Calculate next run
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, _ := parser.Parse(cronExpr)
	nextRun := schedule.Next(parseTime(now)).Format("2006-01-02T15:04:05Z07:00")

	_, err = globalDB.Exec(`
		UPDATE schedules SET last_run = ?, next_run = ?, updated_at = ? WHERE id = ?
	`, now, nextRun, now, scheduleID)
	if err != nil {
		log.Printf("Scheduler: 스케줄 #%d 업데이트 실패: %v", scheduleID, err)
	}
}

// parseTime parses ISO 8601 time string
func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now()
	}
	return t
}

// JobCount returns the number of registered jobs
func JobCount() int {
	if globalScheduler == nil {
		return 0
	}
	globalScheduler.mu.RLock()
	defer globalScheduler.mu.RUnlock()
	return len(globalScheduler.jobs)
}
