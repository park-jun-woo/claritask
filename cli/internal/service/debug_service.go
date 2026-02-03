package service

import (
	"fmt"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// RunInteractiveDebugging starts an interactive debugging session
// Deprecated: Use RunTaskWithTTY instead
func RunInteractiveDebugging(database *db.DB, task *model.Task) error {
	return RunTaskWithTTY(database, task)
}

// VerifyAfterDebugging verifies the fix after debugging
// Deprecated: Use VerifyTask instead
func VerifyAfterDebugging(task *model.Task) (bool, error) {
	return VerifyTask(task)
}

// ExecuteWithFallback executes a task with fallback to interactive mode
func ExecuteWithFallback(database *db.DB, task *model.Task, manifest *model.Manifest) error {
	// Try headless execution first
	result, err := ExecuteTaskWithClaude(task, manifest)
	if err == nil && result.Success {
		return nil
	}

	fmt.Println("[Claritask] Headless execution failed. Switching to interactive mode...")

	// Fallback to interactive mode
	if err := RunTaskWithTTY(database, task); err != nil {
		return fmt.Errorf("interactive execution failed: %w", err)
	}

	// Verify after execution
	passed, verifyErr := VerifyTask(task)
	if !passed {
		return fmt.Errorf("verification failed: %w", verifyErr)
	}

	return nil
}

// MaxDebugAttempts is the maximum number of debugging attempts
const MaxDebugAttempts = 3
