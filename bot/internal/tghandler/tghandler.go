package tghandler

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"parkjunwoo.com/claribot/internal/handler"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/telegram"
)

// buttonPattern matches [name:value] format
var buttonPattern = regexp.MustCompile(`\[([^:\]]+):([^\]]+)\]`)

// cliOnlyCommands are commands that cannot be used via Telegram
var cliOnlyCommands = []string{
	"project add",
}

// Handler handles Telegram messages
type Handler struct {
	bot            *telegram.Bot
	router         *handler.Router
	pendingContext map[int64]string
	mu             sync.RWMutex // protects pendingContext
}

// New creates a new Telegram handler
func New(bot *telegram.Bot, router *handler.Router) *Handler {
	h := &Handler{
		bot:            bot,
		router:         router,
		pendingContext: make(map[int64]string),
	}

	// Register menu commands
	bot.SetCommands([]telegram.Command{
		{Command: "start", Description: "시작"},
		{Command: "status", Description: "현재 상태"},
		{Command: "project", Description: "프로젝트 관리"},
		{Command: "task", Description: "작업 관리"},
		{Command: "edge", Description: "의존성 관리"},
		{Command: "message", Description: "메시지 관리"},
	})

	return h
}

// parseButtons extracts [name:value] patterns and returns clean message + buttons
// Buttons on the same line are grouped into the same row
func parseButtons(msg string) (string, [][]telegram.Button) {
	lines := strings.Split(msg, "\n")
	var cleanLines []string
	var buttons [][]telegram.Button

	for _, line := range lines {
		matches := buttonPattern.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			cleanLines = append(cleanLines, line)
			continue
		}

		// Remove button patterns from line
		cleanLine := buttonPattern.ReplaceAllString(line, "")
		cleanLine = strings.TrimSpace(cleanLine)
		if cleanLine != "" {
			cleanLines = append(cleanLines, cleanLine)
		}

		// Group buttons on same line into one row
		var row []telegram.Button
		for _, match := range matches {
			name := match[1]
			value := match[2]
			row = append(row, telegram.Button{Text: name, Data: "input:" + value})
		}
		if len(row) > 0 {
			buttons = append(buttons, row)
		}
	}

	cleanMsg := strings.TrimSpace(strings.Join(cleanLines, "\n"))
	return cleanMsg, buttons
}

// sendResult sends a result message, converting [name:value] to buttons
func (h *Handler) sendResult(chatID int64, result types.Result) {
	cleanMsg, buttons := parseButtons(result.Message)

	var err error
	if len(buttons) > 0 {
		err = h.bot.SendWithButtons(chatID, cleanMsg, buttons)
	} else {
		err = h.bot.Send(chatID, result.Message)
	}
	if err != nil {
		log.Printf("[Telegram] 메시지 전송 실패: %v", err)
	}

	// Store context if needs input
	if result.NeedsInput && result.Context != "" {
		h.mu.Lock()
		h.pendingContext[chatID] = result.Context
		h.mu.Unlock()
	}
}

// sendReport sends Claude response as rendered HTML (inline or file)
func (h *Handler) sendReport(chatID int64, result types.Result) {
	cleanMsg, buttons := parseButtons(result.Message)

	var err error
	if len(buttons) > 0 {
		err = h.bot.SendReportWithButtons(chatID, cleanMsg, buttons)
	} else {
		err = h.bot.SendReport(chatID, cleanMsg)
	}
	if err != nil {
		log.Printf("[Telegram] Report 전송 실패: %v (길이: %d)", err, len(cleanMsg))
	}

	// Store context if needs input
	if result.NeedsInput && result.Context != "" {
		h.mu.Lock()
		h.pendingContext[chatID] = result.Context
		h.mu.Unlock()
	}
}

// quickCommands are commands that don't require Claude execution (fast response)
var quickCommands = []string{
	"project", "task list", "task get", "edge list", "edge get",
	"message list", "message get", "message status",
	"schedule list", "schedule get", "schedule runs", "schedule run",
	"status",
}

// needsClaudeExecution checks if a command requires Claude execution
func needsClaudeExecution(cmd string) bool {
	for _, quick := range quickCommands {
		if strings.HasPrefix(cmd, quick) {
			return false
		}
	}
	// Commands that run Claude
	claudeCommands := []string{
		"task plan", "task run", "task cycle",
		"message send", "send",
	}
	for _, cc := range claudeCommands {
		if strings.HasPrefix(cmd, cc) {
			return true
		}
	}
	// Default: if not a known command, it goes to handleClaude
	return true
}

// HandleMessage handles incoming Telegram messages
func (h *Handler) HandleMessage(msg telegram.Message) {
	log.Printf("[Telegram] %s: %s", msg.Username, msg.Text)

	// Handle /start specially (welcome message)
	if msg.Text == "/start" {
		h.sendResult(msg.ChatID, types.Result{
			Success: true,
			Message: "Claribot 시작!\n[프로젝트:project list][상태:status]",
		})
		return
	}

	// Handle / commands via router
	if strings.HasPrefix(msg.Text, "/") {
		cmd := strings.TrimPrefix(msg.Text, "/")
		// Replace _ with space for menu commands (e.g., /project_list → project list)
		cmd = strings.ReplaceAll(cmd, "_", " ")

		// Check for CLI-only commands
		if isCLIOnly(cmd) {
			h.bot.Send(msg.ChatID, "이 명령어는 CLI에서만 사용 가능합니다.")
			return
		}

		// Quick commands: synchronous processing
		if !needsClaudeExecution(cmd) {
			result := h.router.Execute(cmd)
			h.sendResult(msg.ChatID, result)
			return
		}

		// Claude commands: async processing
		h.bot.Send(msg.ChatID, "처리 중...")
		go func() {
			result := h.router.Execute(cmd)
			h.sendReport(msg.ChatID, result)
		}()
		return
	}

	// Check for pending context (tikitaka continuation)
	h.mu.Lock()
	ctx, ok := h.pendingContext[msg.ChatID]
	if ok {
		delete(h.pendingContext, msg.ChatID)
	}
	h.mu.Unlock()

	if ok {
		cmd := ctx + " " + msg.Text

		// Quick commands: synchronous
		if !needsClaudeExecution(cmd) {
			result := h.router.Execute(cmd)
			h.sendResult(msg.ChatID, result)
			return
		}

		// Claude commands: async
		h.bot.Send(msg.ChatID, "처리 중...")
		go func() {
			result := h.router.Execute(cmd)
			h.sendReport(msg.ChatID, result)
		}()
		return
	}

	// Handle message with current project context (or global)
	projectID, _ := h.router.GetProject()
	label := projectID
	if label == "" {
		label = "global"
	}

	// Route plain text to "message send" command with telegram source (async)
	h.bot.Send(msg.ChatID, fmt.Sprintf("[%s] 메시지 처리 중...", label))
	go func() {
		result := h.router.Execute("message send telegram " + msg.Text)
		h.sendReport(msg.ChatID, result)
	}()
}

// isCLIOnly checks if a command is CLI-only
func isCLIOnly(cmd string) bool {
	for _, cliCmd := range cliOnlyCommands {
		if strings.HasPrefix(cmd, cliCmd) {
			return true
		}
	}
	return false
}

// HandleCallback handles Telegram callback queries (button clicks)
func (h *Handler) HandleCallback(cb telegram.Callback) {
	log.Printf("[Callback] %s: %s", cb.Username, cb.Data)

	// Handle input buttons (from [name:value] pattern)
	if strings.HasPrefix(cb.Data, "input:") {
		value := strings.TrimPrefix(cb.Data, "input:")
		h.bot.AnswerCallback(cb.ID, value)

		// Check for pending context
		h.mu.Lock()
		ctx, ok := h.pendingContext[cb.ChatID]
		if ok {
			delete(h.pendingContext, cb.ChatID)
		}
		h.mu.Unlock()

		var cmd string
		if ok {
			cmd = ctx + " " + value
		} else {
			cmd = value
		}

		// Check for CLI-only commands
		if isCLIOnly(cmd) {
			h.bot.Send(cb.ChatID, "이 명령어는 CLI에서만 사용 가능합니다.")
			return
		}

		// Quick commands: synchronous
		if !needsClaudeExecution(cmd) {
			result := h.router.Execute(cmd)
			h.sendResult(cb.ChatID, result)
			return
		}

		// Claude commands: async
		h.bot.Send(cb.ChatID, "처리 중...")
		go func() {
			result := h.router.Execute(cmd)
			h.sendReport(cb.ChatID, result)
		}()
		return
	}

	// Handle project switch (quick command, synchronous)
	if strings.HasPrefix(cb.Data, "switch:") {
		projectID := strings.TrimPrefix(cb.Data, "switch:")
		result := h.router.Execute("project switch " + projectID)
		h.bot.AnswerCallback(cb.ID, projectID+" 선택됨")
		h.bot.Send(cb.ChatID, result.Message)
		return
	}

	h.bot.AnswerCallback(cb.ID, "")
}
