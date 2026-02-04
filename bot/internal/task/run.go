package task

import "parkjunwoo.com/claribot/internal/types"

// Run runs a task
func Run(projectPath, id string) types.Result {
	if id == "" {
		return types.Result{
			Success: true,
			Message: "task run (next pending)",
		}
	}
	return types.Result{
		Success: true,
		Message: "task run " + id,
	}
}
