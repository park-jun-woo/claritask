package task

import "parkjunwoo.com/claribot/internal/types"

// Get gets task details
func Get(projectPath, id string) types.Result {
	return types.Result{
		Success: true,
		Message: "task get " + id,
	}
}
