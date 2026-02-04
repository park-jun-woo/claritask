package message

import (
	"fmt"
	"log"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/prompts"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/claude"
)

// Send creates a message, sends it to Claude Code, and returns the report
func Send(projectPath, content, source string) types.Result {
	return SendWithProject(nil, projectPath, content, source)
}

// SendWithProject creates a message with optional project association
func SendWithProject(projectID *string, projectPath, content, source string) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer globalDB.Close()

	// Insert message with pending status
	now := db.TimeNow()
	result, err := globalDB.Exec(`
		INSERT INTO messages (project_id, content, source, status, created_at)
		VALUES (?, ?, ?, 'pending', ?)
	`, projectID, content, source, now)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("메시지 저장 실패: %v", err),
		}
	}

	msgID, _ := result.LastInsertId()

	// Update status to processing
	_, err = globalDB.Exec(`UPDATE messages SET status = 'processing' WHERE id = ?`, msgID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("상태 업데이트 실패: %v", err),
		}
	}

	// Get system prompt template
	systemPrompt, err := prompts.Get(prompts.Common, "message")
	if err != nil {
		systemPrompt = defaultSystemPrompt()
	}

	// Execute Claude Code
	opts := claude.Options{
		UserPrompt:   content,
		SystemPrompt: systemPrompt,
		WorkDir:      projectPath,
	}

	claudeResult, err := claude.Run(opts)
	if err != nil {
		// Update status to failed
		completedAt := db.TimeNow()
		if _, dbErr := globalDB.Exec(`
			UPDATE messages
			SET status = 'failed', error = ?, completed_at = ?
			WHERE id = ?
		`, err.Error(), completedAt, msgID); dbErr != nil {
			log.Printf("[Message] 에러 저장 실패 (msg #%d): %v", msgID, dbErr)
		}

		return types.Result{
			Success: false,
			Message: fmt.Sprintf("Claude 실행 오류: %v", err),
		}
	}

	// Update status to done with result
	completedAt := db.TimeNow()
	_, err = globalDB.Exec(`
		UPDATE messages
		SET status = 'done', result = ?, completed_at = ?
		WHERE id = ?
	`, claudeResult.Output, completedAt, msgID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("결과 저장 실패: %v", err),
		}
	}

	return types.Result{
		Success: claudeResult.ExitCode == 0,
		Message: claudeResult.Output,
		Data: &Message{
			ID:          int(msgID),
			Content:     content,
			Source:      source,
			Status:      "done",
			Result:      claudeResult.Output,
			CreatedAt:   now,
			CompletedAt: &completedAt,
		},
	}
}

func defaultSystemPrompt() string {
	return `당신은 프로젝트 어시스턴트입니다. 사용자의 메시지를 분석하고 요청된 작업을 수행하세요.

작업 완료 후 다음 형식으로 보고서를 작성하세요:

## 요약
- 수행한 작업 간략 설명

## 상세
- 변경사항 또는 결과 상세

## 다음 단계 (선택)
- 추가 필요한 작업이 있다면 제안
`
}
