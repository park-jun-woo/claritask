package project

import "parkjunwoo.com/claribot/internal/types"

// Create creates a new project
func Create(id string) types.Result {
	return types.Result{
		Success: true,
		Message: "project create " + id,
	}
}
