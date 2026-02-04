package claude

import "context"

// Runner defines the interface for Claude Code execution
// This interface allows for mocking in tests
type Runner interface {
	// Run executes Claude Code with the given options
	Run(ctx context.Context, opts Options) (*Result, error)

	// Available returns number of available slots
	Available() int

	// Max returns max concurrent instances
	Max() int

	// ActiveSessions returns number of active sessions
	ActiveSessions() int
}

// Ensure Manager implements Runner
var _ Runner = (*Manager)(nil)
