package message

// Message represents a user message
type Message struct {
	ID          int     `json:"id"`
	ProjectID   *string `json:"project_id,omitempty"`
	Content     string  `json:"content"`
	Source      string  `json:"source"`
	Status      string  `json:"status"`
	Result      string  `json:"result"`
	Error       string  `json:"error,omitempty"`
	CreatedAt   string  `json:"created_at"`
	CompletedAt *string `json:"completed_at,omitempty"`
}
