package task

import "parkjunwoo.com/claribot/internal/types"

// Delete deletes a task
func Delete(projectPath, id string) types.Result {
	return types.Result{
		Success: true,
		Message: "task delete " + id,
	}
}
