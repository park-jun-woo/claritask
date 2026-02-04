package message

// Message represents a user message
type Message struct {
	ID          int
	Content     string
	Source      string
	Status      string
	Result      string
	Error       string
	CreatedAt   string
	CompletedAt *string
}
