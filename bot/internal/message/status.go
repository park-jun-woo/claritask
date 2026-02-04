package message

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// StatusSummary represents message status summary
type StatusSummary struct {
	Total      int `json:"total"`
	Pending    int `json:"pending"`
	Processing int `json:"processing"`
	Done       int `json:"done"`
	Failed     int `json:"failed"`
}

// Status returns message status summary
func Status(projectPath string) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	var summary StatusSummary

	// Count by status
	rows, err := globalDB.Query(`
		SELECT status, COUNT(*) as cnt
		FROM messages
		GROUP BY status
	`)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		summary.Total += count
		switch status {
		case "pending":
			summary.Pending = count
		case "processing":
			summary.Processing = count
		case "done":
			summary.Done = count
		case "failed":
			summary.Failed = count
		}
	}

	var sb strings.Builder
	sb.WriteString("📊 메시지 상태\n")
	sb.WriteString(fmt.Sprintf("  총: %d\n", summary.Total))
	sb.WriteString(fmt.Sprintf("  ⏳ 대기: %d\n", summary.Pending))
	sb.WriteString(fmt.Sprintf("  🔄 처리중: %d\n", summary.Processing))
	sb.WriteString(fmt.Sprintf("  ✅ 완료: %d\n", summary.Done))
	sb.WriteString(fmt.Sprintf("  ❌ 실패: %d\n", summary.Failed))

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    &summary,
	}
}

// Processing returns list of currently processing messages
func Processing(projectPath string) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	rows, err := globalDB.Query(`
		SELECT id, content, source, created_at
		FROM messages
		WHERE status = 'processing'
		ORDER BY id DESC
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
		if err := rows.Scan(&m.ID, &m.Content, &m.Source, &m.CreatedAt); err != nil {
			continue
		}
		m.Status = "processing"
		messages = append(messages, m)
	}

	if len(messages) == 0 {
		return types.Result{
			Success: true,
			Message: "처리 중인 메시지가 없습니다.",
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔄 처리 중 (%d개)\n", len(messages)))
	for _, m := range messages {
		content := m.Content
		if len(content) > 40 {
			content = content[:40] + "..."
		}
		sb.WriteString(fmt.Sprintf("  #%d [%s] %s\n", m.ID, m.Source, content))
		sb.WriteString(fmt.Sprintf("       시작: %s\n", m.CreatedAt))
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    messages,
	}
}
