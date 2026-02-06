package spec

import (
	"database/sql"
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Get gets spec details
func Get(projectPath, id string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var s Spec
	err = localDB.QueryRow(`
		SELECT id, title, content, status, priority, created_at, updated_at
		FROM specs WHERE id = ?
	`, id).Scan(&s.ID, &s.Title, &s.Content, &s.Status, &s.Priority, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("스펙을 찾을 수 없습니다: #%s", id),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}

	statusIcon := statusToIcon(s.Status)
	msg := fmt.Sprintf("%s #%d %s\nStatus: %s\nCreated: %s", statusIcon, s.ID, s.Title, s.Status, s.CreatedAt)
	if s.Priority != 0 {
		msg += fmt.Sprintf("\nPriority: %d", s.Priority)
	}
	if s.Content != "" {
		msg += fmt.Sprintf("\n\n📄 Content:\n%s", s.Content)
	}

	// Action buttons
	msg += fmt.Sprintf("\n\n[상태변경:spec set %s status][삭제:spec delete %s]", id, id)

	return types.Result{
		Success: true,
		Message: msg,
		Data:    &s,
	}
}

func statusToIcon(status string) string {
	switch status {
	case "draft":
		return "📝"
	case "review":
		return "🔍"
	case "approved":
		return "✅"
	case "deprecated":
		return "🗄️"
	default:
		return "•"
	}
}
