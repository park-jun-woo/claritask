package message

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/task"
)

// BuildContextMap builds a context map text combining recent messages and task tree.
// Returns empty string on failure to avoid disrupting message processing.
func BuildContextMap(globalDB *db.DB, projectPath string, projectID *string) string {
	var sb strings.Builder

	// Section 1: Recent messages from global DB
	msgSection := buildMessageSection(globalDB)
	if msgSection != "" {
		sb.WriteString("## 최근 대화 이력\n\n")
		sb.WriteString(msgSection)
		sb.WriteString("\n")
	}

	// Section 2: Task tree from local DB (only if project path exists)
	if projectPath != "" {
		taskSection := buildTaskSection(projectPath)
		if taskSection != "" {
			sb.WriteString("## Task 현황\n\n")
			sb.WriteString("```\n")
			sb.WriteString(taskSection)
			sb.WriteString("```\n\n")
		}
	}

	if sb.Len() == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("# Context Map\n\n")
	result.WriteString(sb.String())
	return result.String()
}

// buildMessageSection queries recent 10 messages and formats them as a summary.
func buildMessageSection(globalDB *db.DB) string {
	rows, err := globalDB.Query(`
		SELECT id, content, COALESCE(result, ''), status, created_at
		FROM messages
		ORDER BY id DESC
		LIMIT 10
	`)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var sb strings.Builder
	count := 0
	for rows.Next() {
		var id int
		var content, result, status, createdAt string
		if err := rows.Scan(&id, &content, &result, &status, &createdAt); err != nil {
			continue
		}

		contentFirst := firstLine(content, 50)
		resultFirst := firstLine(result, 50)

		sb.WriteString(fmt.Sprintf("- #%d [%s] %s", id, status, contentFirst))
		if resultFirst != "" {
			sb.WriteString(fmt.Sprintf(" → %s", resultFirst))
		}
		sb.WriteString("\n")
		count++
	}

	if count == 0 {
		return ""
	}
	return sb.String()
}

// buildTaskSection opens the local DB and builds a task tree summary with stats.
func buildTaskSection(projectPath string) string {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return ""
	}
	defer localDB.Close()

	contextMap, err := task.BuildContextMap(localDB)
	if err != nil {
		return ""
	}

	// Append task stats
	stats := buildTaskStats(localDB)
	if stats != "" {
		return contextMap + stats
	}
	return contextMap
}

// buildTaskStats returns a one-line task status summary.
func buildTaskStats(localDB *db.DB) string {
	rows, err := localDB.Query(`
		SELECT status, COUNT(*) as cnt
		FROM tasks
		GROUP BY status
	`)
	if err != nil {
		return ""
	}
	defer rows.Close()

	statusCounts := make(map[string]int)
	total := 0
	for rows.Next() {
		var status string
		var cnt int
		if err := rows.Scan(&status, &cnt); err != nil {
			continue
		}
		statusCounts[status] = cnt
		total += cnt
	}

	if total == 0 {
		return ""
	}

	var parts []string
	for _, s := range []string{"todo", "planned", "split", "done", "failed"} {
		if c, ok := statusCounts[s]; ok && c > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", s, c))
		}
	}

	return fmt.Sprintf("통계: 총 %d개 (%s)\n", total, strings.Join(parts, ", "))
}

// firstLine returns the first line of text, truncated to maxRunes.
func firstLine(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Take first line only
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)

	if utf8.RuneCountInString(s) > maxRunes {
		s = string([]rune(s)[:maxRunes]) + "..."
	}
	return s
}
