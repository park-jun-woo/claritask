package edge

import "parkjunwoo.com/claribot/internal/types"

// List lists edges
func List(projectPath, taskID string) types.Result {
	if taskID == "" {
		return types.Result{
			Success: true,
			Message: "edge list (all)",
		}
	}
	return types.Result{
		Success: true,
		Message: "edge list " + taskID,
	}
}
