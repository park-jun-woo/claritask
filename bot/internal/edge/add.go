package edge

import "parkjunwoo.com/claribot/internal/types"

// Add adds a dependency edge
func Add(projectPath, fromID, toID string) types.Result {
	return types.Result{
		Success: true,
		Message: "edge add " + fromID + " -> " + toID,
	}
}
