package task

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents YAML metadata in a task .md file
type Frontmatter struct {
	Status   string `yaml:"status"`
	Parent   *int   `yaml:"parent,omitempty"`
	Priority int    `yaml:"priority,omitempty"`
}

// ParseFrontmatter parses a task markdown file content into frontmatter, title, and body.
// Expected format:
//
//	---
//	status: todo
//	---
//	# Title
//
//	Body content
func ParseFrontmatter(content string) (Frontmatter, string, string, error) {
	var fm Frontmatter

	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return fm, "", "", fmt.Errorf("parse frontmatter: missing opening delimiter")
	}

	// Find closing "---"
	rest := content[3:]
	rest = strings.TrimLeft(rest, "\r\n")
	idx := strings.Index(rest, "---")
	if idx < 0 {
		return fm, "", "", fmt.Errorf("parse frontmatter: missing closing delimiter")
	}

	yamlStr := rest[:idx]
	after := rest[idx+3:]
	after = strings.TrimLeft(after, "\r\n")

	if err := yaml.Unmarshal([]byte(yamlStr), &fm); err != nil {
		return fm, "", "", fmt.Errorf("parse frontmatter yaml: %w", err)
	}

	// Extract H1 title
	title, body := extractTitle(after)

	return fm, title, body, nil
}

// extractTitle extracts H1 title from markdown content.
// Returns (title, remaining body).
func extractTitle(content string) (string, string) {
	lines := strings.SplitN(content, "\n", -1)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			title := strings.TrimSpace(trimmed[2:])
			remaining := strings.Join(lines[i+1:], "\n")
			remaining = strings.TrimSpace(remaining)
			return title, remaining
		}
		// Skip empty lines before title
		if trimmed != "" {
			break
		}
	}
	return "", strings.TrimSpace(content)
}

// FormatFrontmatter formats frontmatter, title, and body into a task markdown file content.
func FormatFrontmatter(fm Frontmatter, title, body string) string {
	yamlBytes, _ := yaml.Marshal(&fm)
	yamlStr := strings.TrimSpace(string(yamlBytes))

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(yamlStr)
	sb.WriteString("\n---\n")
	sb.WriteString("# ")
	sb.WriteString(title)
	sb.WriteString("\n")
	if body != "" {
		sb.WriteString("\n")
		sb.WriteString(body)
		sb.WriteString("\n")
	}

	return sb.String()
}
