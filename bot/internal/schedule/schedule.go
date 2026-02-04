package schedule

// Schedule represents a scheduled task
type Schedule struct {
	ID        int
	ProjectID *string // NULL이면 전역
	CronExpr  string
	Message   string
	Enabled   bool
	RunOnce   bool // true면 한 번 실행 후 자동 비활성화
	LastRun   *string
	NextRun   *string
	CreatedAt string
	UpdatedAt string
}

// ScheduleRun represents a schedule execution result
type ScheduleRun struct {
	ID          int
	ScheduleID  int
	Status      string // running, done, failed
	Result      string // Claude Code 실행 결과
	Error       string
	StartedAt   string
	CompletedAt *string
}
