package schedule

// Schedule represents a scheduled task
type Schedule struct {
	ID        int     `json:"id"`
	ProjectID *string `json:"project_id,omitempty"` // NULL이면 전역
	CronExpr  string  `json:"cron_expr"`
	Message   string  `json:"message"`
	Enabled   bool    `json:"enabled"`
	RunOnce   bool    `json:"run_once"` // true면 한 번 실행 후 자동 비활성화
	LastRun   *string `json:"last_run,omitempty"`
	NextRun   *string `json:"next_run,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// ScheduleRun represents a schedule execution result
type ScheduleRun struct {
	ID          int     `json:"id"`
	ScheduleID  int     `json:"schedule_id"`
	Status      string  `json:"status"`       // running, done, failed
	Result      string  `json:"result"`       // Claude Code 실행 결과
	Error       string  `json:"error,omitempty"`
	StartedAt   string  `json:"started_at"`
	CompletedAt *string `json:"completed_at,omitempty"`
}
