package project

import (
	"os"
	"path/filepath"
	"strings"
)

// Project represents a project
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Parallel    int    `json:"parallel"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ProjectType represents a project type with label
type ProjectType struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// ValidTypes is the list of valid project types
var ValidTypes = []ProjectType{
	{ID: "dev.platform", Label: "플랫폼개발"},
	{ID: "dev.cli", Label: "CLI개발"},
	{ID: "write.webnovel", Label: "웹소설집필"},
	{ID: "life.family", Label: "가족생활"},
}

// GetTypeIDs returns just the type IDs for validation
func GetTypeIDs() []string {
	ids := make([]string, len(ValidTypes))
	for i, t := range ValidTypes {
		ids[i] = t.ID
	}
	return ids
}

// FormatTypeList formats the type list for display
func FormatTypeList() string {
	var lines []string
	for _, t := range ValidTypes {
		lines = append(lines, "["+t.Label+":"+t.ID+"]")
	}
	return "  " + strings.Join(lines, "\n  ")
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
