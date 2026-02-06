package edge

// Edge represents a task dependency
type Edge struct {
	FromTaskID int    `json:"from_task_id"`
	ToTaskID   int    `json:"to_task_id"`
	CreatedAt  string `json:"created_at"`
}
