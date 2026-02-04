package task

// Task represents a task
type Task struct {
	ID        int
	ParentID  *int
	Title     string
	Spec      string // 요구사항 명세서 (불변)
	Plan      string // 계획서 (1회차 순회에서 생성)
	Report    string // 완료 보고서 (2회차 순회 후 생성)
	Status    string // spec_ready → plan_ready → done
	Error     string
	CreatedAt string
	UpdatedAt string
}
