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
	"parkjunwoo.com/claribot/pkg/claude"
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
	"usage",
}

// Handler handles Telegram messages
type Handler struct {
	bot            *telegram.Bot
	router         *handler.Router
	allowedUsers   []int64
	pendingContext map[int64]pendingEntry
	mu             sync.RWMutex // protects pendingContext
	sem            chan struct{}

	// Bridge integration
	bridgeManager *claude.BridgeManager
	bridgeEnabled bool
	bridgeChatMap map[string]int64 // projectID → chatID for bridge events
	bridgeMu      sync.RWMutex
}

// New creates a new Telegram handler
func New(bot *telegram.Bot, router *handler.Router, allowedUsers []int64) *Handler {
	h := &Handler{
		bot:            bot,
		router:         router,
		allowedUsers:   allowedUsers,
		pendingContext: make(map[int64]pendingEntry),
		sem:            make(chan struct{}, maxConcurrentGoroutines),
		bridgeChatMap:  make(map[string]int64),
	}

	// Register menu commands
	bot.SetCommands([]telegram.Command{
		{Command: "start", Description: "시작"},
		{Command: "status", Description: "현재 상태"},
		{Command: "project", Description: "프로젝트 관리"},
		{Command: "task", Description: "작업 관리"},
		{Command: "spec", Description: "스펙 관리"},
		{Command: "message", Description: "메시지 관리"},
	})

	return h
}

// SetBridgeManager enables Bridge-based Claude interaction
func (h *Handler) SetBridgeManager(bm *claude.BridgeManager) {
	h.bridgeManager = bm
	h.bridgeEnabled = bm != nil
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
			if !strings.HasPrefix(value, "resume:") && !strings.HasPrefix(value, "switch:") && !strings.HasPrefix(value, "bridge:") {
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
	"spec",
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

	// ---- Bridge-enabled message handling ----
	if h.bridgeEnabled {
		h.handleBridgeMessage(msg)
		return
	}

	// ---- Legacy: Route plain text to "message send" command ----
	projectID, _ := h.router.GetProject()
	label := projectID
	if label == "" {
		label = "global"
	}

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

// handleBridgeMessage sends a plain text message through the Agent Bridge
func (h *Handler) handleBridgeMessage(msg telegram.Message) {
	projectID, projectPath := h.router.GetProject()
	label := projectID
	if label == "" {
		label = "global"
	}

	if projectPath == "" {
		h.bot.Send(msg.ChatID, "프로젝트를 먼저 선택하세요: /project")
		return
	}

	if projectID == "" {
		projectID = "global"
	}

	// Get or create bridge for the project
	bridge, err := h.bridgeManager.GetOrCreate(projectID, projectPath)
	if err != nil {
		log.Printf("[Bridge] Failed to create bridge for %s: %v", projectID, err)
		h.bot.Send(msg.ChatID, fmt.Sprintf("Bridge 시작 실패: %v", err))
		return
	}

	// Map this project to the chat for event delivery
	h.bridgeMu.Lock()
	h.bridgeChatMap[projectID] = msg.ChatID
	h.bridgeMu.Unlock()

	// Set up message handler for this bridge (only once per bridge)
	bridge.SetMessageHandler(func(bmsg claude.BridgeMessage) {
		h.handleBridgeEvent(projectID, bmsg)
	})

	// Send the user message to the bridge
	h.bot.Send(msg.ChatID, fmt.Sprintf("[%s] 🤖 처리 중...", label))
	if err := bridge.SendMessage(msg.Text); err != nil {
		log.Printf("[Bridge] Failed to send message to bridge %s: %v", projectID, err)
		h.bot.Send(msg.ChatID, fmt.Sprintf("메시지 전송 실패: %v", err))
	}
}

// handleBridgeEvent processes events received from the Agent Bridge
func (h *Handler) handleBridgeEvent(projectID string, msg claude.BridgeMessage) {
	h.bridgeMu.RLock()
	chatID, ok := h.bridgeChatMap[projectID]
	h.bridgeMu.RUnlock()

	if !ok {
		log.Printf("[Bridge/%s] No chat mapped for bridge event: %s", projectID, msg.Type)
		return
	}

	switch msg.Type {
	case "init":
		log.Printf("[Bridge/%s] Session initialized: %s", projectID, msg.SessionID)

	case "assistant_text":
		if msg.Content != "" {
			text := fmt.Sprintf("📌 %s\n\n%s", projectID, msg.Content)
			if err := h.bot.SendReport(chatID, text); err != nil {
				log.Printf("[Bridge/%s] Failed to send assistant text: %v", projectID, err)
			}
		}

	case "ask_user":
		h.handleBridgeAskUser(chatID, projectID, msg)

	case "plan_review":
		h.handleBridgePlanReview(chatID, projectID, msg)

	case "tool_request":
		h.handleBridgeToolRequest(chatID, projectID, msg)

	case "result":
		status := "✅"
		if msg.Status == "error" {
			status = "❌"
		}
		text := fmt.Sprintf("📌 %s\n\n%s 완료\n\n%s", projectID, status, msg.Result)
		if msg.CostUSD > 0 {
			text += fmt.Sprintf("\n\n💰 $%.4f", msg.CostUSD)
		}
		if err := h.bot.SendReport(chatID, text); err != nil {
			log.Printf("[Bridge/%s] Failed to send result: %v", projectID, err)
		}

	case "error":
		errText := fmt.Sprintf("📌 %s\n\n❌ 오류: %s", projectID, msg.Message)
		h.bot.Send(chatID, errText)
	}
}

// handleBridgeAskUser sends AskUserQuestion as inline keyboard buttons
func (h *Handler) handleBridgeAskUser(chatID int64, projectID string, msg claude.BridgeMessage) {
	if msg.RequestID == "" || len(msg.Questions) == 0 {
		return
	}

	for _, q := range msg.Questions {
		text := fmt.Sprintf("📌 %s\n\n❓ %s\n%s", projectID, q.Header, q.Question)

		var buttons [][]telegram.Button
		for _, opt := range q.Options {
			// Callback data format: bridge:<requestID>:ask:<question_text>:<label>
			// Truncate to fit Telegram's 64-byte callback data limit
			data := fmt.Sprintf("bridge:%s:ask:%s", msg.RequestID, opt.Label)
			if len(data) > 64 {
				data = data[:64]
			}
			btn := telegram.Button{
				Text: fmt.Sprintf("%s — %s", opt.Label, opt.Description),
				Data: data,
			}
			buttons = append(buttons, []telegram.Button{btn})
		}

		if err := h.bot.SendWithButtons(chatID, text, buttons); err != nil {
			log.Printf("[Bridge/%s] Failed to send ask_user buttons: %v", projectID, err)
		}
	}
}

// handleBridgePlanReview sends plan for review with approve/deny buttons
func (h *Handler) handleBridgePlanReview(chatID int64, projectID string, msg claude.BridgeMessage) {
	if msg.RequestID == "" {
		return
	}

	text := fmt.Sprintf("📌 %s\n\n📋 Plan Review\n\n%s", projectID, msg.Plan)

	approveData := fmt.Sprintf("bridge:%s:plan:approve", msg.RequestID)
	denyData := fmt.Sprintf("bridge:%s:plan:deny", msg.RequestID)

	buttons := [][]telegram.Button{
		{
			{Text: "✅ 승인", Data: approveData},
			{Text: "❌ 거부", Data: denyData},
		},
	}

	if err := h.bot.SendReportWithButtons(chatID, text, buttons); err != nil {
		log.Printf("[Bridge/%s] Failed to send plan review: %v", projectID, err)
	}
}

// handleBridgeToolRequest sends dangerous tool request for approval
func (h *Handler) handleBridgeToolRequest(chatID int64, projectID string, msg claude.BridgeMessage) {
	if msg.RequestID == "" {
		return
	}

	// Show tool details
	var detail string
	if msg.ToolName == "Bash" {
		if cmd, ok := msg.Input["command"].(string); ok {
			detail = fmt.Sprintf("```\n%s\n```", cmd)
		}
	} else {
		detail = fmt.Sprintf("Tool: %s", msg.ToolName)
	}

	text := fmt.Sprintf("📌 %s\n\n⚠️ 위험 명령 승인 요청\n\n%s", projectID, detail)

	allowData := fmt.Sprintf("bridge:%s:tool:allow", msg.RequestID)
	denyData := fmt.Sprintf("bridge:%s:tool:deny", msg.RequestID)

	buttons := [][]telegram.Button{
		{
			{Text: "✅ 허용", Data: allowData},
			{Text: "❌ 거부", Data: denyData},
		},
	}

	if err := h.bot.SendWithButtons(chatID, text, buttons); err != nil {
		log.Printf("[Bridge/%s] Failed to send tool request: %v", projectID, err)
	}
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

	// Handle bridge callbacks (bridge:<requestID>:<type>:<value>)
	if strings.HasPrefix(cb.Data, "bridge:") {
		h.handleBridgeCallback(cb)
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
		snapshot := h.router.SnapshotContext()
		if snapshot.ProjectPath != "" {
			task.ClearCycleState(snapshot.ProjectPath)
		}
		h.bot.Send(cb.ChatID, fmt.Sprintf("🔄 %s 재개합니다...", cmd))
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

// handleBridgeCallback processes bridge-related callback button clicks
// Format: bridge:<requestID>:<type>:<value>
func (h *Handler) handleBridgeCallback(cb telegram.Callback) {
	parts := strings.SplitN(cb.Data, ":", 4)
	if len(parts) < 3 {
		h.bot.AnswerCallback(cb.ID, "잘못된 콜백 데이터")
		return
	}

	requestID := parts[1]
	cbType := parts[2]
	value := ""
	if len(parts) > 3 {
		value = parts[3]
	}

	// Find the bridge that has this requestID
	// We search all active bridges since we don't know which project it belongs to
	h.bridgeMu.RLock()
	var bridge *claude.Bridge
	for pid := range h.bridgeChatMap {
		if b := h.bridgeManager.GetBridge(pid); b != nil {
			bridge = b
			break
		}
	}
	h.bridgeMu.RUnlock()

	if bridge == nil {
		h.bot.AnswerCallback(cb.ID, "Bridge가 실행 중이지 않습니다")
		return
	}

	switch cbType {
	case "ask":
		// AskUserQuestion answer: value is the selected label
		h.bot.AnswerCallback(cb.ID, value)
		// For simplicity, we send the answer for the first question
		// The requestID maps to the full AskUserQuestion call
		answers := map[string]string{}
		// We store the answer with a placeholder key — the bridge will match it
		answers["_selected"] = value
		bridge.RespondToQuestion(requestID, answers, nil)

	case "plan":
		// Plan review: approve or deny
		approve := value == "approve"
		if approve {
			h.bot.AnswerCallback(cb.ID, "계획 승인됨")
		} else {
			h.bot.AnswerCallback(cb.ID, "계획 거부됨")
		}
		bridge.RespondToPlan(requestID, approve)

	case "tool":
		// Tool request: allow or deny
		allow := value == "allow"
		if allow {
			h.bot.AnswerCallback(cb.ID, "명령 허용됨")
			bridge.RespondToTool(requestID, true, map[string]interface{}{}, "")
		} else {
			h.bot.AnswerCallback(cb.ID, "명령 거부됨")
			bridge.RespondToTool(requestID, false, nil, "User denied this action")
		}

	default:
		h.bot.AnswerCallback(cb.ID, "")
	}
}
