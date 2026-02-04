package message

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// List lists recent messages
func List(projectPath string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	rows, err := localDB.Query(`
		SELECT id, content, source, status, created_at
		FROM messages
		ORDER BY id DESC
		LIMIT 20
	`)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Content, &m.Source, &m.Status, &m.CreatedAt); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("스캔 실패: %v", err),
			}
		}
		messages = append(messages, m)
	}

	if len(messages) == 0 {
		return types.Result{
			Success: true,
			Message: "메시지가 없습니다.",
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("메시지 (%d):\n", len(messages)))
	for _, m := range messages {
		statusIcon := statusToIcon(m.Status)
		// Truncate content for display
		content := m.Content
		if len(content) > 30 {
			content = content[:30] + "..."
		}
		sb.WriteString(fmt.Sprintf("  %s [#%d:message get %d] %s\n", statusIcon, m.ID, m.ID, content))
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    messages,
	}
}

func statusToIcon(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "processing":
		return "🔄"
	case "done":
		return "✅"
	case "failed":
		return "❌"
	default:
		return "•"
	}
}
