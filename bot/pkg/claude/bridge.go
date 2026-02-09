package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"
)

// BridgeConfig holds configuration for Agent Bridge
type BridgeConfig struct {
	BridgePath     string        // path to compiled bridge (e.g., bot/bridge/dist/index.js)
	NodePath       string        // path to node binary (default: "node")
	IdleTimeout    time.Duration // bridge process idle timeout (0 = no timeout)
	PermissionMode string        // default, bypassPermissions, acceptEdits, plan
}

// DefaultBridgeConfig returns default bridge configuration
func DefaultBridgeConfig() BridgeConfig {
	return BridgeConfig{
		BridgePath:     "bot/bridge/dist/index.js",
		NodePath:       "node",
		IdleTimeout:    30 * time.Minute,
		PermissionMode: "bypassPermissions",
	}
}

// Bridge represents a running Agent Bridge process for a project
type Bridge struct {
	projectID   string
	projectPath string
	sessionID   string
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	scanner     *bufio.Scanner
	mu          sync.Mutex
	callbacks   map[string]chan BridgeToolResult // requestID → response channel
	callbackMu  sync.Mutex
	onMessage   func(BridgeMessage)
	started     bool
	closed      bool
	readDone    chan struct{} // signals reader goroutine has ended
}

// BridgeManager manages per-project Bridge processes
type BridgeManager struct {
	bridges map[string]*Bridge // projectID → Bridge
	config  BridgeConfig
	mu      sync.RWMutex
}

// NewBridgeManager creates a new BridgeManager
func NewBridgeManager(cfg BridgeConfig) *BridgeManager {
	return &BridgeManager{
		bridges: make(map[string]*Bridge),
		config:  cfg,
	}
}

// GetOrCreate returns an existing Bridge for the project or creates a new one
func (bm *BridgeManager) GetOrCreate(projectID, projectPath string) (*Bridge, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if b, ok := bm.bridges[projectID]; ok && !b.closed {
		return b, nil
	}

	b, err := bm.startBridge(projectID, projectPath)
	if err != nil {
		return nil, err
	}
	bm.bridges[projectID] = b
	return b, nil
}

// GetBridge returns the existing Bridge for a project (nil if not running)
func (bm *BridgeManager) GetBridge(projectID string) *Bridge {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	if b, ok := bm.bridges[projectID]; ok && !b.closed {
		return b
	}
	return nil
}

// startBridge starts a new bridge process
func (bm *BridgeManager) startBridge(projectID, projectPath string) (*Bridge, error) {
	nodePath := bm.config.NodePath
	if nodePath == "" {
		nodePath = "node"
	}

	cmd := exec.Command(nodePath, bm.config.BridgePath)
	cmd.Dir = projectPath

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Stderr goes to log
	cmd.Stderr = &bridgeStderrWriter{projectID: projectID}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to start bridge process: %w", err)
	}

	log.Printf("[Bridge/%s] Process started (pid: %d)", projectID, cmd.Process.Pid)

	b := &Bridge{
		projectID:   projectID,
		projectPath: projectPath,
		cmd:         cmd,
		stdin:       stdin,
		scanner:     bufio.NewScanner(stdout),
		callbacks:   make(map[string]chan BridgeToolResult),
		readDone:    make(chan struct{}),
	}

	// Increase scanner buffer for large messages
	b.scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	// Send start message
	permMode := bm.config.PermissionMode
	if permMode == "" {
		permMode = "bypassPermissions"
	}

	startMsg := BridgeStartMsg{
		Type:           "start",
		ProjectPath:    projectPath,
		PermissionMode: permMode,
	}
	if err := b.writeJSON(startMsg); err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return nil, fmt.Errorf("failed to send start message: %w", err)
	}

	b.started = true

	// Start reading stdout in background
	go b.readLoop()

	// Start process monitor
	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Printf("[Bridge/%s] Process exited: %v", projectID, err)
		} else {
			log.Printf("[Bridge/%s] Process exited normally", projectID)
		}
		b.mu.Lock()
		b.closed = true
		b.mu.Unlock()

		// Remove from manager
		bm.mu.Lock()
		delete(bm.bridges, projectID)
		bm.mu.Unlock()
	}()

	return b, nil
}

// SetMessageHandler sets the callback for receiving bridge messages
func (b *Bridge) SetMessageHandler(fn func(BridgeMessage)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onMessage = fn
}

// SendMessage sends a user message to the bridge
func (b *Bridge) SendMessage(content string) error {
	msg := BridgeUserMsg{
		Type:    "user_message",
		Content: content,
	}
	return b.writeJSON(msg)
}

// RespondToTool responds to a tool_request from the bridge
func (b *Bridge) RespondToTool(requestID string, allow bool, input map[string]interface{}, denyMsg string) error {
	result := BridgeToolResult{}
	if allow {
		result.Behavior = "allow"
		result.UpdatedInput = input
	} else {
		result.Behavior = "deny"
		result.Message = denyMsg
	}

	msg := BridgeToolResponseMsg{
		Type:      "tool_response",
		RequestID: requestID,
		Result:    result,
	}
	return b.writeJSON(msg)
}

// RespondToQuestion responds to an ask_user request with answers
func (b *Bridge) RespondToQuestion(requestID string, answers map[string]string, questions []BridgeQuestion) error {
	updatedInput := map[string]interface{}{
		"answers":   answers,
		"questions": questions,
	}
	return b.RespondToTool(requestID, true, updatedInput, "")
}

// RespondToPlan responds to a plan_review request (approve or deny)
func (b *Bridge) RespondToPlan(requestID string, approve bool) error {
	if approve {
		return b.RespondToTool(requestID, true, map[string]interface{}{}, "")
	}
	return b.RespondToTool(requestID, false, nil, "User rejected the plan")
}

// Interrupt sends an interrupt signal to stop execution
func (b *Bridge) Interrupt() error {
	msg := BridgeInterruptMsg{Type: "interrupt"}
	return b.writeJSON(msg)
}

// SessionID returns the current session ID (set after init message)
func (b *Bridge) SessionID() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sessionID
}

// IsClosed returns whether the bridge process has terminated
func (b *Bridge) IsClosed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.closed
}

// Close terminates the bridge process
func (b *Bridge) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	b.mu.Unlock()

	// Close stdin to signal the bridge to exit
	b.stdin.Close()

	// Wait for reader to finish
	select {
	case <-b.readDone:
	case <-time.After(5 * time.Second):
	}

	// Kill if still running
	if b.cmd.Process != nil {
		b.cmd.Process.Kill()
	}

	return nil
}

// writeJSON writes a JSON message to the bridge's stdin
func (b *Bridge) writeJSON(v interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("bridge is closed")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	data = append(data, '\n')
	_, err = b.stdin.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to bridge stdin: %w", err)
	}

	return nil
}

// readLoop reads NDJSON from bridge stdout and dispatches messages
func (b *Bridge) readLoop() {
	defer close(b.readDone)

	for b.scanner.Scan() {
		line := b.scanner.Text()
		if line == "" {
			continue
		}

		var msg BridgeMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Printf("[Bridge/%s] invalid JSON from bridge: %s", b.projectID, line)
			continue
		}

		// Handle init message to store session ID
		if msg.Type == "init" && msg.SessionID != "" {
			b.mu.Lock()
			b.sessionID = msg.SessionID
			b.mu.Unlock()
			log.Printf("[Bridge/%s] Session initialized: %s", b.projectID, msg.SessionID)
		}

		// Dispatch to handler
		b.mu.Lock()
		handler := b.onMessage
		b.mu.Unlock()

		if handler != nil {
			handler(msg)
		}
	}

	if err := b.scanner.Err(); err != nil {
		log.Printf("[Bridge/%s] Scanner error: %v", b.projectID, err)
	}
}

// Shutdown closes all running bridges
func (bm *BridgeManager) Shutdown() {
	bm.mu.Lock()
	bridges := make([]*Bridge, 0, len(bm.bridges))
	for _, b := range bm.bridges {
		bridges = append(bridges, b)
	}
	bm.mu.Unlock()

	for _, b := range bridges {
		log.Printf("[Bridge/%s] Shutting down...", b.projectID)
		b.Close()
	}
}

// ActiveBridges returns the number of active bridge processes
func (bm *BridgeManager) ActiveBridges() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return len(bm.bridges)
}

// bridgeStderrWriter forwards bridge stderr to Go log
type bridgeStderrWriter struct {
	projectID string
}

func (w *bridgeStderrWriter) Write(p []byte) (n int, err error) {
	log.Printf("[Bridge/%s] %s", w.projectID, string(p))
	return len(p), nil
}
