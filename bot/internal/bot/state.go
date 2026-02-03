package bot

import (
	"sync"
	"time"
)

// WaitingType represents what input the bot is waiting for
type WaitingType int

const (
	WaitingNone WaitingType = iota
	WaitingTaskTitle
	WaitingTaskDescription
	WaitingTaskExpert
	WaitingMessageTitle
	WaitingMessageContent
	WaitingMessageRecipient
	WaitingExpertQuestion
	WaitingConfirmation
)

// UserState holds the conversation state for a user
type UserState struct {
	CurrentProject string
	WaitingFor     WaitingType
	TempData       map[string]interface{}
	UpdatedAt      time.Time
}

// StateManager manages user states
type StateManager struct {
	mu     sync.RWMutex
	states map[int64]*UserState
	ttl    time.Duration
}

// NewStateManager creates a new state manager
func NewStateManager(ttl time.Duration) *StateManager {
	sm := &StateManager{
		states: make(map[int64]*UserState),
		ttl:    ttl,
	}

	// Start cleanup goroutine
	go sm.cleanupLoop()

	return sm
}

// Get returns the state for a user
func (sm *StateManager) Get(userID int64) *UserState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	state, exists := sm.states[userID]
	if !exists {
		return &UserState{
			WaitingFor: WaitingNone,
			TempData:   make(map[string]interface{}),
		}
	}
	return state
}

// Set sets the state for a user
func (sm *StateManager) Set(userID int64, state *UserState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state.UpdatedAt = time.Now()
	sm.states[userID] = state
}

// SetWaiting sets the waiting type for a user
func (sm *StateManager) SetWaiting(userID int64, waiting WaitingType) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state, exists := sm.states[userID]
	if !exists {
		state = &UserState{
			TempData: make(map[string]interface{}),
		}
		sm.states[userID] = state
	}
	state.WaitingFor = waiting
	state.UpdatedAt = time.Now()
}

// SetTempData sets temporary data for a user
func (sm *StateManager) SetTempData(userID int64, key string, value interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state, exists := sm.states[userID]
	if !exists {
		state = &UserState{
			TempData: make(map[string]interface{}),
		}
		sm.states[userID] = state
	}
	if state.TempData == nil {
		state.TempData = make(map[string]interface{})
	}
	state.TempData[key] = value
	state.UpdatedAt = time.Now()
}

// GetTempData gets temporary data for a user
func (sm *StateManager) GetTempData(userID int64, key string) interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	state, exists := sm.states[userID]
	if !exists || state.TempData == nil {
		return nil
	}
	return state.TempData[key]
}

// SetCurrentProject sets the current project for a user
func (sm *StateManager) SetCurrentProject(userID int64, projectID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state, exists := sm.states[userID]
	if !exists {
		state = &UserState{
			TempData: make(map[string]interface{}),
		}
		sm.states[userID] = state
	}
	state.CurrentProject = projectID
	state.UpdatedAt = time.Now()
}

// Clear clears the state for a user
func (sm *StateManager) Clear(userID int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if state, exists := sm.states[userID]; exists {
		state.WaitingFor = WaitingNone
		state.TempData = make(map[string]interface{})
		state.UpdatedAt = time.Now()
	}
}

// cleanupLoop periodically removes expired states
func (sm *StateManager) cleanupLoop() {
	ticker := time.NewTicker(sm.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanup()
	}
}

// cleanup removes expired states
func (sm *StateManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for userID, state := range sm.states {
		if now.Sub(state.UpdatedAt) > sm.ttl {
			delete(sm.states, userID)
		}
	}
}
