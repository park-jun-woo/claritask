package project

import (
	"os"
	"path/filepath"
)

// Project represents a project
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Parallel    int    `json:"parallel"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// DefaultPath is the default project creation path
var DefaultPath string

// SetDefaultPath sets the default project path
func SetDefaultPath(path string) {
	if path != "" {
		// Expand ~ to home directory
		if len(path) > 0 && path[0] == '~' {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, path[1:])
		}
		DefaultPath = path
	}
}
