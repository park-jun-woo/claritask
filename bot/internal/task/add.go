package task

import "parkjunwoo.com/claribot/internal/types"

// Add adds a new task
func Add(projectPath, title string) types.Result {
	return types.Result{
		Success: true,
		Message: "task add: " + title,
	}
}
