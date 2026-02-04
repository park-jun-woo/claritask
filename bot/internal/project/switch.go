package project

import "parkjunwoo.com/claribot/internal/types"

// Switch switches to a project
func Switch(id string) types.Result {
	return types.Result{
		Success: true,
		Message: "project switch " + id,
		Data: &Project{
			ID:   id,
			Path: "/tmp/" + id, // placeholder
		},
	}
}
