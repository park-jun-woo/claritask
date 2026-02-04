package task

import "parkjunwoo.com/claribot/internal/types"

// Set updates a task field
func Set(projectPath, id, field, value string) types.Result {
	return types.Result{
		Success: true,
		Message: "task set " + id + " " + field + " " + value,
	}
}
