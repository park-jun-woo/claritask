package message

import (
	"fmt"
	"strconv"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Get retrieves a message by ID
func Get(projectPath, idStr string) types.Result {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return types.Result{
			Success: false,
			Message: "잘못된 ID 형식",
		}
	}

	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var m Message
	var completedAt *string
	err = localDB.QueryRow(`
		SELECT id, content, source, status, result, error, created_at, completed_at
		FROM messages WHERE id = ?
	`, id).Scan(&m.ID, &m.Content, &m.Source, &m.Status, &m.Result, &m.Error, &m.CreatedAt, &completedAt)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("메시지 조회 실패: %v", err),
		}
	}
	m.CompletedAt = completedAt

	msg := fmt.Sprintf("메시지 #%d\n상태: %s\n소스: %s\n생성: %s\n\n내용:\n%s",
		m.ID, m.Status, m.Source, m.CreatedAt, m.Content)

	if m.Result != "" {
		msg += fmt.Sprintf("\n\n결과:\n%s", m.Result)
	}
	if m.Error != "" {
		msg += fmt.Sprintf("\n\n오류:\n%s", m.Error)
	}

	return types.Result{
		Success: true,
		Message: msg,
		Data:    &m,
	}
}
