package project

import "parkjunwoo.com/claribot/internal/types"

// List lists all projects
func List() types.Result {
	return types.Result{
		Success: true,
		Message: "project list",
	}
}
