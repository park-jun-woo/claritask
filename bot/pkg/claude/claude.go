package claude

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"github.com/creack/pty"
)

// ansiEscapePattern matches ANSI escape sequences
var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]|\x1b\][^\x07]*\x07|\x1b\][^\x1b]*\x1b\\`)

// Config holds global Claude Code execution settings
type Config struct {
	Timeout time.Duration // idle timeout (no output)
	Max     int           // max concurrent instances
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Timeout: 1200 * time.Second, // 20 minutes
		Max:     3,
	}
}

// Manager manages Claude Code execution with concurrency control
type Manager struct {
	config   Config
	sem      chan struct{}
	mu       sync.RWMutex
	sessions map[*Session]struct{} // track active sessions
	closed   bool
}

// global manager instance
var (
	globalManager *Manager
	managerOnce   sync.Once
)

// Init initializes the global manager with config
func Init(cfg Config) {
	managerOnce.Do(func() {
		globalManager = &Manager{
			config:   cfg,
			sem:      make(chan struct{}, cfg.Max),
			sessions: make(map[*Session]struct{}),
		}
	})
}

// Shutdown gracefully shuts down all active sessions
func Shutdown() {
	if globalManager == nil {
		return
	}
	globalManager.Shutdown()
}

// Shutdown closes all active sessions
func (m *Manager) Shutdown() {
	m.mu.Lock()
	m.closed = true
	sessions := make([]*Session, 0, len(m.sessions))
	for s := range m.sessions {
		sessions = append(sessions, s)
	}
	m.mu.Unlock()

	for _, s := range sessions {
		s.Close()
	}
}

// ActiveSessions returns number of active sessions
func ActiveSessions() int {
	if globalManager == nil {
		return 0
	}
	return globalManager.ActiveSessions()
}

// ActiveSessions returns number of active sessions
func (m *Manager) ActiveSessions() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// GetManager returns the global manager, initializing with defaults if needed
func GetManager() *Manager {
	if globalManager == nil {
		Init(DefaultConfig())
	}
	return globalManager
}

// Result contains the response from Claude Code
type Result struct {
	Output   string
	ExitCode int
}

// Options for Claude Code execution
type Options struct {
	SystemPrompt string
	UserPrompt   string
	Model        string        // optional: sonnet, opus, haiku
	WorkDir      string        // working directory
	Timeout      time.Duration // override idle timeout (0 = use config)
	AllowedTools []string      // optional: limit available tools
}

// Run executes Claude Code with PTY and returns the result
// Uses print mode (-p) for single-shot execution
// Blocks if max concurrent instances reached (FIFO queue)
func Run(opts Options) (*Result, error) {
	return RunContext(context.Background(), opts)
}

// RunContext executes Claude Code with context for cancellation
func RunContext(ctx context.Context, opts Options) (*Result, error) {
	mgr := GetManager()
	return mgr.Run(ctx, opts)
}

// Run executes Claude Code with concurrency control
func (m *Manager) Run(ctx context.Context, opts Options) (*Result, error) {
	// Acquire semaphore (blocks if max reached, FIFO order)
	select {
	case m.sem <- struct{}{}:
		// acquired
		log.Printf("[Claude] Semaphore acquired (used: %d/%d)", len(m.sem), m.config.Max)
	case <-ctx.Done():
		return nil, fmt.Errorf("cancelled while waiting in queue: %w", ctx.Err())
	}
	defer func() {
		<-m.sem // release
		log.Printf("[Claude] Semaphore released (used: %d/%d)", len(m.sem), m.config.Max)
	}()

	return m.execute(ctx, opts)
}

// execute runs Claude Code with PTY and idle timeout
func (m *Manager) execute(ctx context.Context, opts Options) (*Result, error) {
	args := buildArgs(opts)

	cmd := exec.CommandContext(ctx, "claude", args...)
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	// Start with PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start claude with pty: %w", err)
	}
	defer ptmx.Close()

	// Determine idle timeout
	idleTimeout := m.config.Timeout
	if opts.Timeout > 0 {
		idleTimeout = opts.Timeout
	}

	// Read output with idle timeout
	output, err := m.readWithIdleTimeout(ctx, ptmx, cmd, idleTimeout)
	if err != nil {
		// Kill process on read error to prevent zombie
		cmd.Process.Kill()
		cmd.Wait()
		return nil, err
	}

	// Wait for completion with timeout (prevent infinite block)
	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	var waitErr error
	select {
	case waitErr = <-waitDone:
		// Process exited normally
	case <-time.After(10 * time.Second):
		// Wait timeout - force kill
		cmd.Process.Kill()
		<-waitDone // drain
		waitErr = fmt.Errorf("wait timeout, process killed")
	}

	// Get exit code
	exitCode := 0
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if ctx.Err() != nil {
			return nil, fmt.Errorf("execution cancelled: %w", ctx.Err())
		}
	}

	return &Result{
		Output:   stripANSI(output),
		ExitCode: exitCode,
	}, nil
}

// readWithIdleTimeout reads from PTY with idle timeout
// Kills process if no output for idleTimeout duration
func (m *Manager) readWithIdleTimeout(ctx context.Context, ptmx *os.File, cmd *exec.Cmd, idleTimeout time.Duration) (string, error) {
	var output bytes.Buffer
	buf := make([]byte, 4096)

	resultCh := make(chan error, 1)

	go func() {
		for {
			// Set read deadline for idle timeout
			ptmx.SetReadDeadline(time.Now().Add(idleTimeout))

			n, err := ptmx.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
			}

			if err != nil {
				if os.IsTimeout(err) {
					// Idle timeout - kill process
					cmd.Process.Kill()
					resultCh <- fmt.Errorf("idle timeout: no output for %v", idleTimeout)
					return
				}
				// EOF or other error - process ended
				resultCh <- nil
				return
			}
		}
	}()

	select {
	case err := <-resultCh:
		if err != nil {
			return output.String(), err
		}
		return output.String(), nil
	case <-ctx.Done():
		cmd.Process.Kill()
		return output.String(), ctx.Err()
	}
}

// QueueLength returns current number of waiting executions
func (m *Manager) QueueLength() int {
	return len(m.sem)
}

// Available returns number of available slots
func (m *Manager) Available() int {
	return m.config.Max - len(m.sem)
}

// Max returns max concurrent instances
func (m *Manager) Max() int {
	return m.config.Max
}

// Status returns a status summary
type Status struct {
	Max       int `json:"max"`
	Used      int `json:"used"`
	Available int `json:"available"`
	Sessions  int `json:"sessions"`
}

// GetStatus returns global manager status
func GetStatus() Status {
	mgr := GetManager()
	return Status{
		Max:       mgr.config.Max,
		Used:      len(mgr.sem),
		Available: mgr.config.Max - len(mgr.sem),
		Sessions:  mgr.ActiveSessions(),
	}
}

// Session represents an interactive Claude Code session
type Session struct {
	cmd     *exec.Cmd
	pty     *os.File
	cancel  context.CancelFunc
	manager *Manager
	mu      sync.Mutex
}

// StartSession starts an interactive Claude Code session
// Blocks if max concurrent instances reached
func StartSession(opts Options) (*Session, error) {
	return StartSessionContext(context.Background(), opts)
}

// StartSessionContext starts an interactive session with context
func StartSessionContext(ctx context.Context, opts Options) (*Session, error) {
	mgr := GetManager()
	return mgr.StartSession(ctx, opts)
}

// StartSession starts an interactive Claude Code session with concurrency control
func (m *Manager) StartSession(ctx context.Context, opts Options) (*Session, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return nil, fmt.Errorf("manager is shutting down")
	}
	m.mu.RUnlock()

	// Acquire semaphore
	select {
	case m.sem <- struct{}{}:
		// acquired
	case <-ctx.Done():
		return nil, fmt.Errorf("cancelled while waiting in queue: %w", ctx.Err())
	}

	sessionCtx, cancel := context.WithCancel(context.Background())

	args := buildInteractiveArgs(opts)
	cmd := exec.CommandContext(sessionCtx, "claude", args...)
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		cancel()
		<-m.sem // release
		return nil, fmt.Errorf("failed to start interactive session: %w", err)
	}

	session := &Session{
		cmd:     cmd,
		pty:     ptmx,
		cancel:  cancel,
		manager: m,
	}

	// Track session
	m.mu.Lock()
	m.sessions[session] = struct{}{}
	m.mu.Unlock()

	return session, nil
}

// Send sends a message to the Claude Code session and reads response
func (s *Session) Send(message string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Write message
	_, err := s.pty.Write([]byte(message + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	// Read response with timeout
	var output bytes.Buffer
	buf := make([]byte, 4096)

	// Use manager's idle timeout
	idleTimeout := s.manager.config.Timeout

	for {
		s.pty.SetReadDeadline(time.Now().Add(idleTimeout))

		n, err := s.pty.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
		}

		if err != nil {
			if os.IsTimeout(err) {
				// Check if we have any output - if so, probably response complete
				if output.Len() > 0 {
					break
				}
				return "", fmt.Errorf("idle timeout waiting for response")
			}
			break
		}

		// Reset deadline on each read (only timeout if no output)
		s.pty.SetReadDeadline(time.Now().Add(2 * time.Second))
	}

	return stripANSI(output.String()), nil
}

// Close terminates the Claude Code session and releases semaphore
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancel()
	s.pty.Close()
	err := s.cmd.Wait()

	// Untrack session
	s.manager.mu.Lock()
	delete(s.manager.sessions, s)
	s.manager.mu.Unlock()

	// Release semaphore
	<-s.manager.sem

	return err
}

// stripANSI removes ANSI escape sequences from output
func stripANSI(s string) string {
	return ansiEscapePattern.ReplaceAllString(s, "")
}

// buildArgs constructs command line arguments for print mode
func buildArgs(opts Options) []string {
	args := []string{"-p", "--dangerously-skip-permissions"} // print mode, skip permission prompts

	if opts.SystemPrompt != "" {
		args = append(args, "--system-prompt", opts.SystemPrompt)
	}

	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	if len(opts.AllowedTools) > 0 {
		args = append(args, "--tools")
		args = append(args, opts.AllowedTools...)
	}

	if opts.UserPrompt != "" {
		args = append(args, opts.UserPrompt)
	}

	return args
}

// buildInteractiveArgs constructs command line arguments for interactive mode
func buildInteractiveArgs(opts Options) []string {
	var args []string

	if opts.SystemPrompt != "" {
		args = append(args, "--system-prompt", opts.SystemPrompt)
	}

	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	if len(opts.AllowedTools) > 0 {
		args = append(args, "--tools")
		args = append(args, opts.AllowedTools...)
	}

	if opts.UserPrompt != "" {
		args = append(args, opts.UserPrompt)
	}

	return args
}
