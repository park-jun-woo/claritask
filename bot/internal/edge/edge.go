package edge

// Edge represents a task dependency
type Edge struct {
	FromTaskID int
	ToTaskID   int
	CreatedAt  string
}
