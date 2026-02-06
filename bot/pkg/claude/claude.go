package claude

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/creack/pty"
)

// ansiEscapePattern matches ANSI escape sequences
var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]|\x1b\][^\x07]*\x07|\x1b\][^\x1b]*\x1b\\`)

// Config holds global Claude Code execution settings
type Config struct {
	Timeout    time.Duration // idle timeout (no output)
	MaxTimeout time.Duration // absolute timeout (total execution time)
	Max        int           // max concurrent instances
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Timeout:    1200 * time.Second, // 20 minutes idle
		MaxTimeout: 1800 * time.Second, // 30 minutes absolute
		Max:        3,
	}
}

// Manager manages Claude Code execution with concurrency control
type Manager struct {
	config   Config
	sem      chan struct{}
	mu       sync.RWMutex
	sessions map[*Session]struct{} // track active sessions
	closed   bool
	stopCh   chan struct{}   // signal to stop watchdog
	wg       sync.WaitGroup // wait for watchdog goroutine
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
			stopCh:   make(chan struct{}),
		}
		globalManager.wg.Add(1)
		go globalManager.watchdog()
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
	// Stop watchdog goroutine
	close(m.stopCh)
	m.wg.Wait()

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
	ReportPath   string        // optional: report file path for completion detection
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

// execute runs Claude Code with PTY and idle timeout + absolute timeout
func (m *Manager) execute(ctx context.Context, opts Options) (*Result, error) {
	args := buildArgs(opts)

	// Apply absolute timeout to prevent infinite execution
	absTimeout := m.config.MaxTimeout
	if absTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, absTimeout)
		defer cancel()
		log.Printf("[Claude] Absolute timeout set: %v", absTimeout)
	}

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

	// If ReportPath is set, watch for the report file
	if opts.ReportPath != "" {
		return m.executeWithReportWatch(ctx, ptmx, cmd, opts.ReportPath, idleTimeout)
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

// executeWithReportWatch runs Claude and watches for a report file to detect completion.
// When the report file is created, it reads the file content, kills the Claude process, and returns.
func (m *Manager) executeWithReportWatch(ctx context.Context, ptmx *os.File, cmd *exec.Cmd, reportPath string, idleTimeout time.Duration) (*Result, error) {
	log.Printf("[Claude] Watching for report file: %s", reportPath)

	// Remove report file if it exists from previous run
	os.Remove(reportPath)

	// Start reading PTY output in background (to keep PTY alive)
	var output bytes.Buffer
	ptyDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 4096)
		for {
			ptmx.SetReadDeadline(time.Now().Add(idleTimeout))
			n, err := ptmx.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
			}
			if err != nil {
				if os.IsTimeout(err) {
					cmd.Process.Kill()
					ptyDone <- fmt.Errorf("idle timeout: no output for %v", idleTimeout)
					return
				}
				ptyDone <- nil
				return
			}
		}
	}()

	// Watch for report file
	reportCh := make(chan string, 1)
	stopWatch := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if data, err := os.ReadFile(reportPath); err == nil && len(data) > 0 {
					reportCh <- string(data)
					return
				}
			case <-stopWatch:
				return
			}
		}
	}()

	// Wait for: report file, process exit, or context cancellation
	select {
	case reportContent := <-reportCh:
		close(stopWatch)
		log.Printf("[Claude] Report file detected: %s (%d bytes)", reportPath, len(reportContent))
		// Give Claude a moment to finish writing
		time.Sleep(500 * time.Millisecond)
		// Re-read to ensure complete content
		if data, err := os.ReadFile(reportPath); err == nil && len(data) > 0 {
			reportContent = string(data)
		}
		// Kill the process
		cmd.Process.Kill()
		cmd.Wait()
		return &Result{
			Output:   reportContent,
			ExitCode: 0,
		}, nil

	case err := <-ptyDone:
		close(stopWatch)
		if err != nil {
			return nil, err
		}
		// Process ended - check if report file was created
		if data, readErr := os.ReadFile(reportPath); readErr == nil && len(data) > 0 {
			log.Printf("[Claude] Report file found after process exit: %s", reportPath)
			return &Result{
				Output:   string(data),
				ExitCode: 0,
			}, nil
		}
		// No report file - fall back to PTY output
		log.Printf("[Claude] No report file, using PTY output")

		// Wait for cmd exit
		waitDone := make(chan error, 1)
		go func() { waitDone <- cmd.Wait() }()
		var waitErr error
		select {
		case waitErr = <-waitDone:
		case <-time.After(10 * time.Second):
			cmd.Process.Kill()
			<-waitDone
			waitErr = fmt.Errorf("wait timeout")
		}
		exitCode := 0
		if waitErr != nil {
			if exitErr, ok := waitErr.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
		return &Result{
			Output:   stripANSI(output.String()),
			ExitCode: exitCode,
		}, nil

	case <-ctx.Done():
		close(stopWatch)
		cmd.Process.Kill()
		cmd.Wait()
		return nil, fmt.Errorf("execution cancelled: %w", ctx.Err())
	}
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
	cmd          *exec.Cmd
	pty          *os.File
	cancel       context.CancelFunc
	manager      *Manager
	mu           sync.Mutex
	lastActivity time.Time
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

	// Apply MaxTimeout: use context.WithTimeout if configured, else context.WithCancel
	var sessionCtx context.Context
	var cancel context.CancelFunc
	if m.config.MaxTimeout > 0 {
		sessionCtx, cancel = context.WithTimeout(context.Background(), m.config.MaxTimeout)
		log.Printf("[Claude] Session max timeout set: %v", m.config.MaxTimeout)
	} else {
		sessionCtx, cancel = context.WithCancel(context.Background())
	}

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
		cmd:          cmd,
		pty:          ptmx,
		cancel:       cancel,
		manager:      m,
		lastActivity: time.Now(),
	}

	// Track session
	m.mu.Lock()
	m.sessions[session] = struct{}{}
	m.mu.Unlock()

	// Auto-close session when MaxTimeout is reached
	if m.config.MaxTimeout > 0 {
		go func() {
			<-sessionCtx.Done()
			if sessionCtx.Err() == context.DeadlineExceeded {
				log.Printf("[Claude] Session max timeout reached (%v), auto-closing", m.config.MaxTimeout)
				session.Close()
			}
		}()
	}

	return session, nil
}

// Send sends a message to the Claude Code session and reads response
func (s *Session) Send(message string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastActivity = time.Now()

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

// watchdog periodically checks active sessions for dead processes
func (m *Manager) watchdog() {
	defer m.wg.Done()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Printf("[Claude] Watchdog started (interval: 1m)")

	for {
		select {
		case <-m.stopCh:
			log.Printf("[Claude] Watchdog stopped")
			return
		case <-ticker.C:
			m.mu.RLock()
			sessions := make([]*Session, 0, len(m.sessions))
			for s := range m.sessions {
				sessions = append(sessions, s)
			}
			m.mu.RUnlock()

			for _, s := range sessions {
				if s.cmd.Process == nil {
					continue
				}
				// Signal(0) checks if process is alive without sending a signal
				err := s.cmd.Process.Signal(syscall.Signal(0))
				if err != nil {
					log.Printf("[Claude] Watchdog: process %d is dead (last activity: %s), closing session",
						s.cmd.Process.Pid, s.lastActivity.Format("15:04:05"))
					s.Close()
				}
			}
		}
	}
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

// authErrorPattern matches authentication/authorization error messages from Claude Code
var authErrorPattern = regexp.MustCompile(`(?i)(does not have access|please login again|authentication failed|unauthorized|token expired|invalid.{0,20}credentials|API key.{0,20}(invalid|expired|missing)|not authenticated|session expired|access denied)`)

// IsAuthError checks if a Claude Code result indicates an authentication error
func IsAuthError(result *Result) bool {
	if result == nil || result.ExitCode == 0 {
		return false
	}
	return authErrorPattern.MatchString(result.Output)
}

// stripANSI removes ANSI escape sequences and ensures valid UTF-8 output
func stripANSI(s string) string {
	// Remove ANSI escape sequences
	s = ansiEscapePattern.ReplaceAllString(s, "")

	// Remove other common PTY control characters (ESC, BEL, etc.)
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return r // keep standard whitespace
		}
		if r < 0x20 && r != '\n' && r != '\r' && r != '\t' {
			return -1 // remove control characters
		}
		if r == utf8.RuneError {
			return -1 // remove invalid runes
		}
		return r
	}, s)

	// Force valid UTF-8 encoding (remove any remaining invalid bytes)
	s = strings.ToValidUTF8(s, "")

	return s
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
