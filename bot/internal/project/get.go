package project

import "parkjunwoo.com/claribot/internal/types"

// Get gets project details
func Get(id string) types.Result {
	return types.Result{
		Success: true,
		Message: "project get " + id,
	}
}
