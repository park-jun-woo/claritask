package task

import (
	"fmt"
	"os"
	"strings"
)

// ValidStatus defines allowed status values (matches DB CHECK constraint)
var ValidStatus = map[string]bool{
	"todo":    true,
	"split":   true,
	"planned": true,
	"done":    true,
	"failed":  true,
}

// ValidationError contains file validation results.
type ValidationError struct {
	FilePath string
	Errors   []string // blocking errors
	Warnings []string // non-blocking warnings (e.g. empty body)
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validate %s: %s", e.FilePath, strings.Join(e.Errors, "; "))
}

// ValidateTaskFile validates a task .md file format.
// Returns nil if valid, *ValidationError if there are blocking errors.
func ValidateTaskFile(filePath string) error {
	ve := &ValidationError{FilePath: filePath}

	// 1. File exists and is readable
	data, err := os.ReadFile(filePath)
	if err != nil {
		ve.Errors = append(ve.Errors, fmt.Sprintf("파일 읽기 실패: %v", err))
		return ve
	}

	content := string(data)

	// 2. Frontmatter delimiter exists
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		ve.Errors = append(ve.Errors, "frontmatter 구분자(---) 없음")
		return ve
	}

	fm, title, body, err := ParseFrontmatter(content)
	if err != nil {
		ve.Errors = append(ve.Errors, fmt.Sprintf("frontmatter 파싱 실패: %v", err))
		return ve
	}

	// 3. Status required and must be valid
	if fm.Status == "" {
		ve.Errors = append(ve.Errors, "status 필드 필수")
	} else if !ValidStatus[fm.Status] {
		keys := make([]string, 0, len(ValidStatus))
		for k := range ValidStatus {
			keys = append(keys, k)
		}
		ve.Errors = append(ve.Errors, fmt.Sprintf("허용되지 않는 status: %s (허용: %s)", fm.Status, strings.Join(keys, ", ")))
	}

	// 4. Parent must be positive if present
	if fm.Parent != nil && *fm.Parent <= 0 {
		ve.Errors = append(ve.Errors, fmt.Sprintf("parent는 양수여야 합니다: %d", *fm.Parent))
	}

	// 5. H1 title required
	if title == "" {
		ve.Errors = append(ve.Errors, "H1 제목(# Title) 필수")
	}

	// 6. Empty body is a warning (not an error)
	if strings.TrimSpace(body) == "" {
		ve.Warnings = append(ve.Warnings, "body가 비어있습니다")
	}

	if len(ve.Errors) > 0 {
		return ve
	}

	return nil
}
