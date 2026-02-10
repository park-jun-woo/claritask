package message

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"parkjunwoo.com/claribot/internal/config"
	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/prompts"
	"parkjunwoo.com/claribot/internal/terminal"
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

	msgID, err := result.LastInsertId()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("메시지 ID 획득 실패: %v", err),
		}
	}

	// Update status to processing
	_, err = globalDB.Exec(`UPDATE messages SET status = 'processing' WHERE id = ?`, msgID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("상태 업데이트 실패: %v", err),
		}
	}

	// Build report path
	reportPath := filepath.Join(projectPath, ".claribot", fmt.Sprintf("message-%d-report.md", msgID))
	// Ensure .claribot directory exists
	if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("report 디렉토리 생성 실패: %v", err),
		}
	}

	// Load config for contextMax
	cfg, _ := config.Load()
	contextMax := cfg.Claude.ContextMax

	// Build context map (best-effort, empty string on failure)
	contextMap := BuildContextMap(globalDB, projectPath, projectID, contextMax)

	// Get system prompt template
	systemPrompt, err := prompts.Get("message")
	if err != nil {
		systemPrompt = defaultSystemPrompt()
	}
	systemPrompt = renderPrompt(systemPrompt, map[string]string{
		"ReportPath": reportPath,
		"ContextMap": contextMap,
	})

	// Execute Claude Code
	opts := claude.Options{
		UserPrompt:   content,
		SystemPrompt: systemPrompt,
		WorkDir:      projectPath,
		ReportPath:   reportPath,
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
		// Clean up report file
		if rmErr := os.Remove(reportPath); rmErr != nil && !os.IsNotExist(rmErr) {
			log.Printf("[Message] report 파일 삭제 실패 (msg #%d): %v", msgID, rmErr)
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

	// Clean up report file after DB save
	if rmErr := os.Remove(reportPath); rmErr != nil && !os.IsNotExist(rmErr) {
		log.Printf("[Message] report 파일 삭제 실패 (msg #%d): %v", msgID, rmErr)
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

// renderPrompt renders a Go template string with the given data
func renderPrompt(tmplStr string, data map[string]string) string {
	tmpl, err := template.New("prompt").Parse(tmplStr)
	if err != nil {
		return tmplStr
	}

	// Convert map to struct-like data
	type PromptData struct {
		ReportPath string
		ContextMap string
	}
	d := PromptData{
		ReportPath: data["ReportPath"],
		ContextMap: data["ContextMap"],
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		return tmplStr
	}
	return buf.String()
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

// SendViaPTY creates a message in DB and injects it into an existing PTY session.
// Returns immediately with processing status; a background goroutine polls for the report file.
func SendViaPTY(projectID *string, projectPath, content, source string, session *terminal.PTYSession) types.Result {
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

	msgID, err := result.LastInsertId()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("메시지 ID 획득 실패: %v", err),
		}
	}

	// Update status to processing
	_, err = globalDB.Exec(`UPDATE messages SET status = 'processing' WHERE id = ?`, msgID)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("상태 업데이트 실패: %v", err),
		}
	}

	// Build report path
	reportPath := filepath.Join(projectPath, ".claribot", fmt.Sprintf("message-%d-report.md", msgID))
	if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("report 디렉토리 생성 실패: %v", err),
		}
	}

	// Build prompt and inject into PTY stdin
	prompt := buildPTYPrompt(content, reportPath)
	if err := session.WriteToPTY([]byte(prompt + "\n")); err != nil {
		updateMessageStatus(msgID, "failed", "", fmt.Sprintf("PTY 쓰기 실패: %v", err))
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("PTY 쓰기 실패: %v", err),
		}
	}

	// Start async report file polling
	go pollReportFile(msgID, reportPath, projectID)

	return types.Result{
		Success: true,
		Message: "메시지 전송됨 (PTY)",
		Data: &Message{
			ID:        int(msgID),
			ProjectID: projectID,
			Content:   content,
			Source:    source,
			Status:    "processing",
			CreatedAt: now,
		},
	}
}

// buildPTYPrompt combines user content with report path instruction.
func buildPTYPrompt(content, reportPath string) string {
	return fmt.Sprintf("%s\n\n작업 완료 후 반드시 보고서를 다음 파일에 저장하세요: %s\n보고서 형식: ## 요약 / ## 상세 / ## 다음 단계(선택)", content, reportPath)
}

// pollReportFile polls for the report file and updates DB when found.
func pollReportFile(msgID int64, reportPath string, projectID *string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(10 * time.Minute)

	for {
		select {
		case <-timeout:
			updateMessageStatus(msgID, "failed", "", "report 파일 생성 타임아웃")
			return
		case <-ticker.C:
			data, err := os.ReadFile(reportPath)
			if err != nil {
				continue // file not yet created
			}
			if len(data) == 0 {
				continue // file exists but still being written
			}
			// Report found → update DB
			updateMessageStatus(msgID, "done", string(data), "")
			os.Remove(reportPath)
			return
		}
	}
}

// updateMessageStatus updates message status in the global DB.
func updateMessageStatus(msgID int64, status, resultText, errMsg string) {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		log.Printf("[Message] DB 열기 실패 (msg #%d): %v", msgID, err)
		return
	}
	defer globalDB.Close()

	completedAt := db.TimeNow()
	if errMsg != "" {
		globalDB.Exec(`UPDATE messages SET status = ?, error = ?, completed_at = ? WHERE id = ?`,
			status, errMsg, completedAt, msgID)
	} else {
		globalDB.Exec(`UPDATE messages SET status = ?, result = ?, completed_at = ? WHERE id = ?`,
			status, resultText, completedAt, msgID)
	}
}
