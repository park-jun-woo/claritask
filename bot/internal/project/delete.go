package project

import "parkjunwoo.com/claribot/internal/types"

// Delete deletes a project
func Delete(id string) types.Result {
	return types.Result{
		Success: true,
		Message: "project delete " + id,
	}
}
