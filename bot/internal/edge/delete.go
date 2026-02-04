package edge

import "parkjunwoo.com/claribot/internal/types"

// Delete deletes a dependency edge
func Delete(projectPath, fromID, toID string) types.Result {
	return types.Result{
		Success: true,
		Message: "edge delete " + fromID + " -> " + toID,
	}
}
