package message

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/pagination"
)

// List lists messages with pagination
func List(projectPath string, req pagination.PageRequest) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	// Count total
	var total int
	if err := globalDB.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&total); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("카운트 실패: %v", err),
		}
	}

	if total == 0 {
		return types.Result{
			Success: true,
			Message: "메시지가 없습니다.",
		}
	}

	rows, err := globalDB.Query(`
		SELECT id, content, source, status, created_at
		FROM messages
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, req.Limit(), req.Offset())
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
	if err := rows.Err(); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("행 순회 오류: %v", err),
		}
	}

	pageResp := pagination.NewPageResponse(messages, req.Page, req.PageSize, total)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 메시지 (%d/%d 페이지, 총 %d개)\n", pageResp.Page, pageResp.TotalPages, total))
	for _, m := range messages {
		statusIcon := statusToIcon(m.Status)
		// Truncate content for display
		content := m.Content
		if utf8.RuneCountInString(content) > 30 {
			content = string([]rune(content)[:30]) + "..."
		}
		sb.WriteString(fmt.Sprintf("  %s [#%d:message get %d] %s\n", statusIcon, m.ID, m.ID, content))
	}

	// Add pagination buttons
	if pageResp.HasPrev || pageResp.HasNext {
		sb.WriteString("\n")
		if pageResp.HasPrev {
			sb.WriteString(fmt.Sprintf("[◀ 이전:message list -p %d]", pageResp.Page-1))
		}
		if pageResp.HasNext {
			sb.WriteString(fmt.Sprintf("[다음 ▶:message list -p %d]", pageResp.Page+1))
		}
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    pageResp,
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
