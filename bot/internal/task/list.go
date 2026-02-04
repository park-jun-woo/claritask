package task

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// List lists all tasks (top-level only, parent_id IS NULL)
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
		SELECT id, title, status, created_at
		FROM tasks
		WHERE parent_id IS NULL
		ORDER BY id DESC
	`)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.CreatedAt); err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("스캔 실패: %v", err),
			}
		}
		tasks = append(tasks, t)
	}

	if len(tasks) == 0 {
		return types.Result{
			Success: true,
			Message: "작업이 없습니다.\n[추가:task add]",
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("작업 (%d):\n", len(tasks)))
	for _, t := range tasks {
		statusIcon := statusToIcon(t.Status)
		sb.WriteString(fmt.Sprintf("  %s [#%d:task get %d] %s\n", statusIcon, t.ID, t.ID, t.Title))
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    tasks,
	}
}

func statusToIcon(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "running":
		return "🔄"
	case "done":
		return "✅"
	case "failed":
		return "❌"
	default:
		return "•"
	}
}
