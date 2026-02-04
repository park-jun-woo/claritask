package types

import "strings"

// Request represents a CLI request to the service
type Request struct {
	Command string   `json:"command"`          // e.g., "project", "task", "send"
	Args    []string `json:"args,omitempty"`   // command arguments
	Context string   `json:"context,omitempty"` // for interactive continuation
}

// ToCommandString converts request to command string for router
func (r *Request) ToCommandString() string {
	if len(r.Args) == 0 {
		return r.Command
	}
	result := r.Command
	for _, arg := range r.Args {
		// Quote args that contain spaces
		if strings.Contains(arg, " ") {
			result += " \"" + arg + "\""
		} else {
			result += " " + arg
		}
	}
	return result
}
