package tghandler

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"parkjunwoo.com/claribot/internal/handler"
	"parkjunwoo.com/claribot/internal/task"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/telegram"
)

const (
	maxConcurrentGoroutines = 5
	pendingContextTTL       = 5 * time.Minute
)

// pendingEntry stores a pending context with its creation timestamp
type pendingEntry struct {
	context   string
	createdAt time.Time
}

// safeGo runs a function in a goroutine with panic recovery, error notification, and concurrency limiting
func (h *Handler) safeGo(chatID int64, fn func()) {
	go func() {
		// Acquire semaphore
		select {
		case h.sem <- struct{}{}:
			// acquired
		default:
			log.Printf("[Telegram] goroutine 제한 초과, 대기 중 (chatID: %d)", chatID)
			h.sem <- struct{}{} // block until slot available
		}
		defer func() {
			<-h.sem // release semaphore
		}()

		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Telegram] PANIC in goroutine: %v", r)
				h.bot.Send(chatID, fmt.Sprintf("내부 오류 발생: %v", r))
			}
		}()
		fn()
	}()
}

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
	allowedUsers   []int64
	pendingContext map[int64]pendingEntry
	mu             sync.RWMutex // protects pendingContext
	sem            chan struct{}
}

// New creates a new Telegram handler
func New(bot *telegram.Bot, router *handler.Router, allowedUsers []int64) *Handler {
	h := &Handler{
		bot:            bot,
		router:         router,
		allowedUsers:   allowedUsers,
		pendingContext: make(map[int64]pendingEntry),
		sem:            make(chan struct{}, maxConcurrentGoroutines),
	}

	// Register menu commands
	bot.SetCommands([]telegram.Command{
		{Command: "start", Description: "시작"},
		{Command: "status", Description: "현재 상태"},
		{Command: "project", Description: "프로젝트 관리"},
		{Command: "task", Description: "작업 관리"},
		{Command: "message", Description: "메시지 관리"},
		{Command: "usage", Description: "Claude 사용량"},
	})

	return h
}

// isAllowed checks if a chat ID is in the allowed users list.
// Returns true if allowedUsers is empty (allow all) or chatID is in the list.
func (h *Handler) isAllowed(chatID int64) bool {
	if len(h.allowedUsers) == 0 {
		return true
	}
	for _, id := range h.allowedUsers {
		if id == chatID {
			return true
		}
	}
	return false
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
			// Use value directly if it already has a known callback prefix
			data := value
			if !strings.HasPrefix(value, "resume:") && !strings.HasPrefix(value, "switch:") {
				data = "input:" + value
			}
			row = append(row, telegram.Button{Text: name, Data: data})
		}
		if len(row) > 0 {
			buttons = append(buttons, row)
		}
	}

	cleanMsg := strings.TrimSpace(strings.Join(cleanLines, "\n"))
	return cleanMsg, buttons
}

// projectStatusHeader returns the current project status header line
func (h *Handler) projectStatusHeader() string {
	projectID, _ := h.router.GetProject()
	if projectID != "" {
		return "📌 " + projectID
	}
	return "📌 글로벌 모드"
}

// sendResult sends a result message, converting [name:value] to buttons
func (h *Handler) sendResult(chatID int64, result types.Result) {
	cleanMsg, buttons := parseButtons(result.Message)
	cleanMsg = h.projectStatusHeader() + "\n\n" + cleanMsg

	var err error
	if len(buttons) > 0 {
		err = h.bot.SendWithButtons(chatID, cleanMsg, buttons)
	} else {
		err = h.bot.SendReport(chatID, cleanMsg)
	}
	if err != nil {
		log.Printf("[Telegram] 메시지 전송 실패: %v", err)
	}

	// Store context if needs input
	if result.NeedsInput && result.Context != "" {
		h.mu.Lock()
		h.pendingContext[chatID] = pendingEntry{context: result.Context, createdAt: time.Now()}
		h.mu.Unlock()
	}
}

// sendReport sends Claude response as rendered HTML (inline or file)
func (h *Handler) sendReport(chatID int64, result types.Result) {
	cleanMsg, buttons := parseButtons(result.Message)
	cleanMsg = h.projectStatusHeader() + "\n\n" + cleanMsg

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
		h.pendingContext[chatID] = pendingEntry{context: result.Context, createdAt: time.Now()}
		h.mu.Unlock()
	}
}

// cleanExpiredContexts removes expired pending contexts. Must be called with h.mu held.
func (h *Handler) cleanExpiredContexts() {
	now := time.Now()
	for chatID, entry := range h.pendingContext {
		if now.Sub(entry.createdAt) > pendingContextTTL {
			delete(h.pendingContext, chatID)
		}
	}
}

// quickCommands are commands that don't require Claude execution (fast response)
var quickCommands = []string{
	"project", "task list", "task get", "task stop",
	"message list", "message get", "message status",
	"schedule list", "schedule get", "schedule runs", "schedule run",
	"status",
	"usage",
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

	if !h.isAllowed(msg.ChatID) {
		log.Printf("[Telegram] Unauthorized access from chat ID %d (%s)", msg.ChatID, msg.Username)
		h.bot.Send(msg.ChatID, "인증되지 않은 사용자입니다.")
		return
	}

	// Handle /start specially (welcome message)
	if msg.Text == "/start" {
		h.sendResult(msg.ChatID, types.Result{
			Success: true,
			Message: "Claribot 시작!\n[프로젝트:project list][상태:status]",
		})
		// Set persistent reply keyboard with frequently used commands
		h.bot.SetKeyboard(msg.ChatID, "자주 쓰는 명령어:", [][]string{
			{"/status", "/task_list"},
			{"/task_plan_--all", "/task_run_--all", "/task_cycle"},
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
		snapshot := h.router.SnapshotContext()
		if !needsClaudeExecution(cmd) {
			result := h.router.Execute(snapshot, cmd)
			h.sendResult(msg.ChatID, result)
			return
		}

		// Claude commands: async processing
		procMsgID, _ := h.bot.SendAndGetID(msg.ChatID, "처리 중...")
		h.safeGo(msg.ChatID, func() {
			result := h.router.Execute(snapshot, cmd)
			if procMsgID != 0 {
				h.bot.DeleteMessage(msg.ChatID, procMsgID)
			}
			h.sendReport(msg.ChatID, result)
		})
		return
	}

	// Check for pending context (tikitaka continuation)
	h.mu.Lock()
	entry, ok := h.pendingContext[msg.ChatID]
	if ok {
		delete(h.pendingContext, msg.ChatID)
		// Expire if TTL exceeded
		if time.Since(entry.createdAt) > pendingContextTTL {
			ok = false
		}
	}
	h.cleanExpiredContexts()
	h.mu.Unlock()

	if ok {
		cmd := entry.context + " " + msg.Text
		snapshot := h.router.SnapshotContext()

		// Quick commands: synchronous
		if !needsClaudeExecution(cmd) {
			result := h.router.Execute(snapshot, cmd)
			h.sendResult(msg.ChatID, result)
			return
		}

		// Claude commands: async
		procMsgID, _ := h.bot.SendAndGetID(msg.ChatID, "처리 중...")
		h.safeGo(msg.ChatID, func() {
			result := h.router.Execute(snapshot, cmd)
			if procMsgID != 0 {
				h.bot.DeleteMessage(msg.ChatID, procMsgID)
			}
			h.sendReport(msg.ChatID, result)
		})
		return
	}

	// Handle message with current project context (or global)
	projectID, _ := h.router.GetProject()
	label := projectID
	if label == "" {
		label = "global"
	}

	// Route plain text to "message send" command with telegram source (async)
	snapshot := h.router.SnapshotContext()
	procMsgID, _ := h.bot.SendAndGetID(msg.ChatID, fmt.Sprintf("[%s] 메시지 처리 중...", label))
	h.safeGo(msg.ChatID, func() {
		result := h.router.Execute(snapshot, "message send telegram "+msg.Text)
		if procMsgID != 0 {
			h.bot.DeleteMessage(msg.ChatID, procMsgID)
		}
		h.sendReport(msg.ChatID, result)
	})
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

	if !h.isAllowed(cb.ChatID) {
		log.Printf("[Telegram] Unauthorized callback from chat ID %d (%s)", cb.ChatID, cb.Username)
		h.bot.AnswerCallback(cb.ID, "인증되지 않은 사용자입니다.")
		return
	}

	// Handle input buttons (from [name:value] pattern)
	if strings.HasPrefix(cb.Data, "input:") {
		value := strings.TrimPrefix(cb.Data, "input:")
		h.bot.AnswerCallback(cb.ID, value)

		// Check for pending context
		h.mu.Lock()
		entry, ok := h.pendingContext[cb.ChatID]
		if ok {
			delete(h.pendingContext, cb.ChatID)
			if time.Since(entry.createdAt) > pendingContextTTL {
				ok = false
			}
		}
		h.cleanExpiredContexts()
		h.mu.Unlock()

		var cmd string
		if ok {
			cmd = entry.context + " " + value
		} else {
			cmd = value
		}

		// Check for CLI-only commands
		if isCLIOnly(cmd) {
			h.bot.Send(cb.ChatID, "이 명령어는 CLI에서만 사용 가능합니다.")
			return
		}

		// Quick commands: synchronous
		snapshot := h.router.SnapshotContext()
		if !needsClaudeExecution(cmd) {
			result := h.router.Execute(snapshot, cmd)
			h.sendResult(cb.ChatID, result)
			return
		}

		// Claude commands: async
		procMsgID, _ := h.bot.SendAndGetID(cb.ChatID, "처리 중...")
		h.safeGo(cb.ChatID, func() {
			result := h.router.Execute(snapshot, cmd)
			if procMsgID != 0 {
				h.bot.DeleteMessage(cb.ChatID, procMsgID)
			}
			h.sendReport(cb.ChatID, result)
		})
		return
	}

	// Handle cycle resume
	if strings.HasPrefix(cb.Data, "resume:") {
		cycleType := strings.TrimPrefix(cb.Data, "resume:")
		h.bot.AnswerCallback(cb.ID, "순회 재개 중...")

		// Map cycle type to command
		var cmd string
		switch cycleType {
		case "cycle":
			cmd = "task cycle"
		case "plan":
			cmd = "task plan --all"
		case "run":
			cmd = "task run --all"
		default:
			h.bot.Send(cb.ChatID, fmt.Sprintf("알 수 없는 순회 타입: %s", cycleType))
			return
		}

		// Clear interrupted state and re-execute
		task.ClearCycleState()
		h.bot.Send(cb.ChatID, fmt.Sprintf("🔄 %s 재개합니다...", cmd))

		snapshot := h.router.SnapshotContext()
		h.safeGo(cb.ChatID, func() {
			result := h.router.Execute(snapshot, cmd)
			h.sendReport(cb.ChatID, result)
		})
		return
	}

	// Handle project switch (quick command, synchronous)
	if strings.HasPrefix(cb.Data, "switch:") {
		projectID := strings.TrimPrefix(cb.Data, "switch:")
		snapshot := h.router.SnapshotContext()
		result := h.router.Execute(snapshot, "project switch "+projectID)
		h.bot.AnswerCallback(cb.ID, projectID+" 선택됨")
		h.bot.Send(cb.ChatID, result.Message)
		return
	}

	h.bot.AnswerCallback(cb.ID, "")
}
