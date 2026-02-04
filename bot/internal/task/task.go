package task

// Task represents a task
type Task struct {
	ID          int
	ParentID    *int
	Source      string
	Title       string
	Content     string
	Status      string
	Result      string
	Error       string
	CreatedAt   string
	StartedAt   *string
	CompletedAt *string
}
