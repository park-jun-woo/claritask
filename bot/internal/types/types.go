package types

// Result represents a command execution result
type Result struct {
	Success bool
	Message string
	Data    interface{}
}
