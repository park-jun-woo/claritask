package task

import "parkjunwoo.com/claribot/internal/types"

// List lists all tasks
func List(projectPath string) types.Result {
	return types.Result{
		Success: true,
		Message: "task list",
	}
}
