package task

import (
	"regexp"
	"strings"
)

// PlanResult represents the result of 1회차 순회
type PlanResult struct {
	Type     string  `json:"type"`               // "split" or "planned"
	Plan     string  `json:"plan,omitempty"`      // Plan content (if planned)
	Children []Child `json:"children,omitempty"`  // Child tasks (if split)
}

// Child represents a child task created during subdivision
type Child struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// stripCodeBlocks removes wrapping code block markers (```) from output.
// Claude often wraps structured output in code blocks, which breaks HasPrefix checks.
// Handles: ```\n...\n```, ```text\n...\n```, ```bash\n...\n``` etc.
func stripCodeBlocks(output string) string {
	re := regexp.MustCompile("(?s)^```[a-zA-Z]*\\s*\\n(.*)\\n```\\s*$")
	if m := re.FindStringSubmatch(output); len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	return output
}

// extractMarker finds [SPLIT] or [PLANNED] marker anywhere in the output,
// handling cases where Claude adds preamble text before the marker.
func extractMarker(output string) (marker string, content string, found bool) {
	for _, m := range []string{"[SPLIT]", "[PLANNED]"} {
		idx := strings.Index(output, m)
		if idx >= 0 {
			return m, output[idx:], true
		}
	}
	return "", "", false
}

// ParsePlanOutput parses Claude output from 1회차 순회
func ParsePlanOutput(output string) PlanResult {
	output = strings.TrimSpace(output)

	// Strip code block wrappers (Claude often wraps output in ```)
	output = stripCodeBlocks(output)

	// Try direct prefix match first (fast path)
	if strings.HasPrefix(output, "[SPLIT]") {
		return PlanResult{
			Type:     "split",
			Children: parseSplitChildren(output),
		}
	}

	if strings.HasPrefix(output, "[PLANNED]") {
		plan := strings.TrimPrefix(output, "[PLANNED]")
		plan = strings.TrimSpace(plan)
		return PlanResult{
			Type: "planned",
			Plan: plan,
		}
	}

	// Fallback: search for markers anywhere in output
	// (handles preamble text, \r\n line endings, etc.)
	marker, content, found := extractMarker(output)
	if found {
		switch marker {
		case "[SPLIT]":
			return PlanResult{
				Type:     "split",
				Children: parseSplitChildren(content),
			}
		case "[PLANNED]":
			plan := strings.TrimPrefix(content, "[PLANNED]")
			plan = strings.TrimSpace(plan)
			return PlanResult{
				Type: "planned",
				Plan: plan,
			}
		}
	}

	// Default: treat entire output as plan (backward compatibility)
	return PlanResult{
		Type: "planned",
		Plan: output,
	}
}

// parseSplitChildren extracts child task info from [SPLIT] output
// Format: - Task #<id>: <title>
func parseSplitChildren(output string) []Child {
	var children []Child

	// Pattern: - Task #123: Some title
	re := regexp.MustCompile(`-\s*Task\s*#(\d+):\s*(.+)`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			var id int
			if _, err := regexp.MatchString(`^\d+$`, match[1]); err == nil {
				// Parse ID
				for _, c := range match[1] {
					id = id*10 + int(c-'0')
				}
			}
			children = append(children, Child{
				ID:    id,
				Title: strings.TrimSpace(match[2]),
			})
		}
	}

	return children
}
