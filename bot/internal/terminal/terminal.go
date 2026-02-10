package terminal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"parkjunwoo.com/claribot/pkg/logger"
)

// --- RingBuffer ---

// RingBuffer is a fixed-size circular buffer that retains the most recent data.
type RingBuffer struct {
	buf  []byte
	size int
	w    int  // next write position
	full bool // whether buffer has wrapped
	mu   sync.Mutex
}

// NewRingBuffer creates a ring buffer of the given size.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buf:  make([]byte, size),
		size: size,
	}
}

// Write appends data to the ring buffer, overwriting oldest data if full.
func (rb *RingBuffer) Write(p []byte) (int, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	n := len(p)
	if n >= rb.size {
		// Data larger than buffer: keep only the last 'size' bytes
		copy(rb.buf, p[n-rb.size:])
		rb.w = 0
		rb.full = true
		return n, nil
	}
	// Write data, possibly wrapping around
	space := rb.size - rb.w
	if n <= space {
		copy(rb.buf[rb.w:], p)
	} else {
		copy(rb.buf[rb.w:], p[:space])
		copy(rb.buf, p[space:])
	}
	rb.w = (rb.w + n) % rb.size
	if !rb.full && rb.w <= (rb.w-n+rb.size)%rb.size {
		rb.full = true
	}
	return n, nil
}

// Bytes returns the buffered data in chronological order (oldest first).
func (rb *RingBuffer) Bytes() []byte {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if !rb.full {
		out := make([]byte, rb.w)
		copy(out, rb.buf[:rb.w])
		return out
	}
	// Buffer has wrapped: [w..size) + [0..w)
	out := make([]byte, rb.size)
	copy(out, rb.buf[rb.w:])
	copy(out[rb.size-rb.w:], rb.buf[:rb.w])
	return out
}

// --- controlMessage ---

type controlMessage struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// --- WSBridge (short-lived WebSocket wrapper) ---

// WSBridge represents a single WebSocket connection to a PTYSession.
// All writes go through WriteMessage which holds writeMu to prevent
// concurrent writes (gorilla/websocket is not safe for concurrent writes).
type WSBridge struct {
	ws      *websocket.Conn
	writeMu sync.Mutex
	done    chan struct{}
	once    sync.Once
}

func newWSBridge(ws *websocket.Conn) *WSBridge {
	return &WSBridge{
		ws:   ws,
		done: make(chan struct{}),
	}
}

// WriteMessage sends a message to the WebSocket, serializing concurrent writes.
func (b *WSBridge) WriteMessage(messageType int, data []byte) error {
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	return b.ws.WriteMessage(messageType, data)
}

func (b *WSBridge) Close() {
	b.once.Do(func() {
		close(b.done)
		b.ws.Close()
	})
}

func (b *WSBridge) IsDone() bool {
	select {
	case <-b.done:
		return true
	default:
		return false
	}
}

// --- PTYSession (long-lived) ---

// PTYSession represents a persistent PTY process that outlives WebSocket connections.
type PTYSession struct {
	Key        string
	pty        *os.File
	cmd        *exec.Cmd
	buf        *RingBuffer
	done       chan struct{}
	closeOnce  sync.Once
	bridge     *WSBridge
	bridgeMu   sync.Mutex
	lastDetach time.Time
	workDir    string
	createdAt  time.Time
}

// SessionInfo is the JSON-serializable session info for API responses.
type SessionInfo struct {
	Key       string `json:"key"`
	WorkDir   string `json:"work_dir"`
	CreatedAt string `json:"created_at"`
	Connected bool   `json:"connected"`
	IdleSec   int    `json:"idle_sec,omitempty"`
}

// IsDone returns true if the PTY process has exited.
func (s *PTYSession) IsDone() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

// Close kills the PTY process and cleans up.
func (s *PTYSession) Close() {
	s.closeOnce.Do(func() {
		close(s.done)
		// Detach any connected WS
		s.bridgeMu.Lock()
		if s.bridge != nil {
			s.bridge.Close()
			s.bridge = nil
		}
		s.bridgeMu.Unlock()
		s.pty.Close()
		if s.cmd.Process != nil {
			s.cmd.Process.Kill()
			s.cmd.Wait()
		}
		logger.Info("[terminal] session %s closed", s.Key)
	})
}

// readPTYLoop continuously reads PTY output into the ring buffer
// and forwards to the attached WS bridge if one exists.
func (s *PTYSession) readPTYLoop() {
	buf := make([]byte, 4096)
	for {
		n, err := s.pty.Read(buf)
		if err != nil {
			if err != io.EOF {
				select {
				case <-s.done:
				default:
					logger.Debug("[terminal] %s pty read error: %v", s.Key, err)
				}
			}
			s.Close()
			return
		}
		data := buf[:n]

		// Always write to ring buffer
		s.buf.Write(data)

		// Forward to WS bridge if attached
		s.bridgeMu.Lock()
		b := s.bridge
		s.bridgeMu.Unlock()
		if b != nil && !b.IsDone() {
			if err := b.WriteMessage(websocket.BinaryMessage, data); err != nil {
				logger.Debug("[terminal] %s ws write error during pty read: %v", s.Key, err)
				b.Close()
				s.bridgeMu.Lock()
				if s.bridge == b {
					s.bridge = nil
					s.lastDetach = time.Now()
				}
				s.bridgeMu.Unlock()
			}
		}
	}
}

// Attach connects a WebSocket to this PTY session.
// It replays the ring buffer and starts reader/keepalive goroutines.
// Returns the number of bytes replayed.
func (s *PTYSession) Attach(ws *websocket.Conn) (int, error) {
	if s.IsDone() {
		return 0, fmt.Errorf("session is closed")
	}

	bridge := newWSBridge(ws)

	// Replace bridge. Old bridge's goroutines will exit on next write/read error.
	s.bridgeMu.Lock()
	s.bridge = bridge
	s.bridgeMu.Unlock()

	// Replay ring buffer
	replay := s.buf.Bytes()
	replayed := len(replay)
	if replayed > 0 {
		if err := bridge.WriteMessage(websocket.BinaryMessage, replay); err != nil {
			bridge.Close()
			s.bridgeMu.Lock()
			if s.bridge == bridge {
				s.bridge = nil
				s.lastDetach = time.Now()
			}
			s.bridgeMu.Unlock()
			return 0, fmt.Errorf("replay failed: %w", err)
		}
	}

	// WS reader goroutine: reads from WS and writes to PTY
	go s.wsReaderLoop(bridge)

	// Keepalive goroutine: sends WS pings
	go s.keepaliveLoop(bridge)

	return replayed, nil
}

// Detach disconnects the current WS bridge without killing the PTY.
func (s *PTYSession) Detach() {
	s.bridgeMu.Lock()
	defer s.bridgeMu.Unlock()
	if s.bridge != nil {
		s.bridge.Close()
		s.bridge = nil
		s.lastDetach = time.Now()
		logger.Info("[terminal] %s detached (PTY alive)", s.Key)
	}
}

// wsReaderLoop reads messages from the WebSocket and writes to the PTY.
func (s *PTYSession) wsReaderLoop(b *WSBridge) {
	// Set initial read deadline and pong handler
	b.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	b.ws.SetPongHandler(func(appData string) error {
		b.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		msgType, data, err := b.ws.ReadMessage()
		if err != nil {
			// WS read error → detach (not close!)
			logger.Debug("[terminal] %s ws read error: %v", s.Key, err)
			s.bridgeMu.Lock()
			if s.bridge == b {
				s.bridge = nil
				s.lastDetach = time.Now()
			}
			s.bridgeMu.Unlock()
			b.Close()
			return
		}

		// Reset read deadline
		b.ws.SetReadDeadline(time.Now().Add(60 * time.Second))

		switch msgType {
		case websocket.BinaryMessage:
			if _, err := s.pty.Write(data); err != nil {
				logger.Debug("[terminal] %s pty write error: %v", s.Key, err)
				s.Close()
				return
			}
		case websocket.TextMessage:
			var msg controlMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			switch msg.Type {
			case "resize":
				if msg.Cols > 0 && msg.Rows > 0 {
					pty.Setsize(s.pty, &pty.Winsize{Cols: msg.Cols, Rows: msg.Rows})
				}
			case "ping":
				b.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
			case "close":
				// Explicit close request → kill PTY
				logger.Info("[terminal] %s received close request", s.Key)
				s.Close()
				return
			}
		}
	}
}

// keepaliveLoop sends WebSocket pings periodically.
func (s *PTYSession) keepaliveLoop(b *WSBridge) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-b.done:
			return
		case <-s.done:
			return
		case <-ticker.C:
			if err := b.WriteMessage(websocket.PingMessage, nil); err != nil {
				s.bridgeMu.Lock()
				if s.bridge == b {
					s.bridge = nil
					s.lastDetach = time.Now()
				}
				s.bridgeMu.Unlock()
				b.Close()
				return
			}
		}
	}
}

// SendMessage sends a message through the current WS bridge (thread-safe).
func (s *PTYSession) SendMessage(messageType int, data []byte) error {
	s.bridgeMu.Lock()
	b := s.bridge
	s.bridgeMu.Unlock()
	if b == nil || b.IsDone() {
		return fmt.Errorf("no active bridge")
	}
	return b.WriteMessage(messageType, data)
}

// WriteToPTY writes data to the PTY stdin (for external callers like message send).
func (s *PTYSession) WriteToPTY(data []byte) error {
	if s.IsDone() {
		return fmt.Errorf("session is closed")
	}
	_, err := s.pty.Write(data)
	return err
}

// Info returns session info for API responses.
func (s *PTYSession) Info() SessionInfo {
	s.bridgeMu.Lock()
	connected := s.bridge != nil && !s.bridge.IsDone()
	s.bridgeMu.Unlock()

	info := SessionInfo{
		Key:       s.Key,
		WorkDir:   s.workDir,
		CreatedAt: s.createdAt.Format(time.RFC3339),
		Connected: connected,
	}
	if !connected && !s.lastDetach.IsZero() {
		info.IdleSec = int(time.Since(s.lastDetach).Seconds())
	}
	return info
}

// --- Manager ---

// Manager manages terminal sessions keyed by project ID or "__global__".
type Manager struct {
	sessions    map[string]*PTYSession
	mu          sync.Mutex
	maxSessions int
	idleTimeout time.Duration
	stopGC      chan struct{}
}

// NewManager creates a new terminal session manager.
// idleTimeout controls how long a detached session lives before being reaped.
func NewManager(max int, idleTimeout time.Duration) *Manager {
	if max <= 0 {
		max = 5
	}
	if idleTimeout <= 0 {
		idleTimeout = 30 * time.Minute
	}
	m := &Manager{
		sessions:    make(map[string]*PTYSession),
		maxSessions: max,
		idleTimeout: idleTimeout,
		stopGC:      make(chan struct{}),
	}
	go m.gc()
	return m
}

// gc periodically cleans dead and idle sessions.
func (m *Manager) gc() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopGC:
			return
		case <-ticker.C:
			m.cleanDeadAndIdle()
		}
	}
}

// cleanDeadAndIdle removes sessions that have exited or exceeded idle timeout.
func (m *Manager) cleanDeadAndIdle() {
	m.mu.Lock()
	var dead []*PTYSession
	now := time.Now()
	for key, s := range m.sessions {
		// Check if process exited
		if s.cmd.ProcessState != nil || s.IsDone() {
			dead = append(dead, s)
			delete(m.sessions, key)
			logger.Info("[terminal] gc: session %s process exited", key)
			continue
		}
		// Check idle timeout (only for detached sessions)
		s.bridgeMu.Lock()
		connected := s.bridge != nil && !s.bridge.IsDone()
		lastDetach := s.lastDetach
		s.bridgeMu.Unlock()

		if !connected && !lastDetach.IsZero() && now.Sub(lastDetach) > m.idleTimeout {
			dead = append(dead, s)
			delete(m.sessions, key)
			logger.Info("[terminal] gc: session %s idle timeout (%.0fs)", key, now.Sub(lastDetach).Seconds())
		}
	}
	m.mu.Unlock()

	for _, s := range dead {
		s.Close()
	}
}

// GetSession returns an existing live session by key, or nil if not found or done.
func (m *Manager) GetSession(key string) *PTYSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[key]
	if !ok || s.IsDone() {
		return nil
	}
	return s
}

// GetOrCreate returns an existing live session or creates a new one.
// Returns (session, isNew, error).
func (m *Manager) GetOrCreate(key string, cols, rows uint16, workDir, initialCmd string) (*PTYSession, bool, error) {
	m.mu.Lock()

	// Check for existing session
	if s, ok := m.sessions[key]; ok {
		if !s.IsDone() && s.cmd.ProcessState == nil {
			m.mu.Unlock()
			return s, false, nil
		}
		// Dead session — remove and create new
		delete(m.sessions, key)
		m.mu.Unlock()
		s.Close()
		m.mu.Lock()
	}

	// Check capacity
	if len(m.sessions) >= m.maxSessions {
		m.mu.Unlock()
		return nil, false, fmt.Errorf("max sessions (%d) reached", m.maxSessions)
	}
	m.mu.Unlock()

	// Create new PTY
	cmd := exec.Command("bash")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	if workDir != "" {
		cmd.Dir = workDir
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: cols, Rows: rows})
	if err != nil {
		return nil, false, fmt.Errorf("pty start: %w", err)
	}

	session := &PTYSession{
		Key:       key,
		pty:       ptmx,
		cmd:       cmd,
		buf:       NewRingBuffer(64 * 1024), // 64KB ring buffer
		done:      make(chan struct{}),
		workDir:   workDir,
		createdAt: time.Now(),
	}

	// Send initial command
	if initialCmd != "" {
		go func() {
			time.Sleep(200 * time.Millisecond)
			ptmx.Write([]byte(initialCmd + "\n"))
		}()
	}

	// Start PTY reader loop (writes to ring buffer + WS bridge)
	go session.readPTYLoop()

	m.mu.Lock()
	m.sessions[key] = session
	m.mu.Unlock()

	logger.Info("[terminal] session %s created (cols=%d, rows=%d, dir=%s)", key, cols, rows, workDir)
	return session, true, nil
}

// Remove explicitly removes and closes a session.
func (m *Manager) Remove(key string) {
	m.mu.Lock()
	s, ok := m.sessions[key]
	if ok {
		delete(m.sessions, key)
	}
	m.mu.Unlock()
	if ok {
		s.Close()
		logger.Info("[terminal] session %s removed (active: %d)", key, m.ActiveCount())
	}
}

// ListSessions returns info about all active sessions.
func (m *Manager) ListSessions() []SessionInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	var infos []SessionInfo
	for _, s := range m.sessions {
		infos = append(infos, s.Info())
	}
	return infos
}

// CloseAll terminates all active sessions.
func (m *Manager) CloseAll() {
	close(m.stopGC)

	m.mu.Lock()
	sessions := make([]*PTYSession, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	m.sessions = make(map[string]*PTYSession)
	m.mu.Unlock()

	for _, s := range sessions {
		s.Close()
	}
}

// ActiveCount returns the number of active sessions.
func (m *Manager) ActiveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sessions)
}
