package task

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"parkjunwoo.com/claribot/pkg/claude"
)

// CycleState represents the current state of a cycle/plan/run operation
type CycleState struct {
	Running       bool      `json:"running"`
	Type          string    `json:"type"`            // "cycle", "plan", "run"
	StartedAt     time.Time `json:"started_at"`
	CurrentTaskID int       `json:"current_task_id"`
	ProjectPath   string    `json:"project_path"`
	ActiveWorkers int       `json:"active_workers"`
}

// activeWorkers tracks the number of currently running parallel workers
var activeWorkers atomic.Int32

var (
	cycleMu           sync.RWMutex
	currentCycleState CycleState
	cycleCancel       context.CancelFunc // cancel function for the current cycle's context
)

// SetCycleState sets the current cycle state
func SetCycleState(state CycleState) {
	cycleMu.Lock()
	defer cycleMu.Unlock()
	currentCycleState = state
}

// SetCycleCancel stores the cancel function for the current cycle's context
func SetCycleCancel(cancel context.CancelFunc) {
	cycleMu.Lock()
	defer cycleMu.Unlock()
	cycleCancel = cancel
}

// CancelCycle calls the stored cancel function to immediately cancel the running cycle context
func CancelCycle() {
	cycleMu.Lock()
	cancel := cycleCancel
	cycleMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// ClearCycleState resets the cycle state to empty
func ClearCycleState() {
	cycleMu.Lock()
	defer cycleMu.Unlock()
	currentCycleState = CycleState{}
	cycleCancel = nil
}

// GetCycleState returns a copy of the current cycle state
func GetCycleState() CycleState {
	cycleMu.RLock()
	defer cycleMu.RUnlock()
	return currentCycleState
}

// UpdateCurrentTask updates only the CurrentTaskID field
func UpdateCurrentTask(taskID int) {
	cycleMu.Lock()
	defer cycleMu.Unlock()
	currentCycleState.CurrentTaskID = taskID
}

// UpdateActiveWorkers adjusts the active worker count by delta (+1 or -1)
func UpdateActiveWorkers(delta int) {
	activeWorkers.Add(int32(delta))
	cycleMu.Lock()
	currentCycleState.ActiveWorkers = int(activeWorkers.Load())
	cycleMu.Unlock()
}

// GetActiveWorkers returns the current active worker count
func GetActiveWorkers() int {
	return int(activeWorkers.Load())
}

// ResetActiveWorkers resets the active worker count to zero
func ResetActiveWorkers() {
	activeWorkers.Store(0)
}

// IsCycleInterrupted returns true if a cycle was started but Claude is no longer running
func IsCycleInterrupted() bool {
	state := GetCycleState()
	if !state.Running {
		return false
	}
	status := claude.GetStatus()
	return status.Used == 0
}

// CycleStatusInfo represents the current cycle status for display
type CycleStatusInfo struct {
	Status        string    // "running", "interrupted", "idle"
	Type          string    // "cycle", "plan", "run"
	StartedAt     time.Time
	CurrentTaskID int
	ProjectPath   string
	ActiveWorkers int
}

// GetCycleStatus returns the current cycle status by cross-checking CycleState and Claude semaphore
func GetCycleStatus() CycleStatusInfo {
	state := GetCycleState()

	if !state.Running {
		return CycleStatusInfo{Status: "idle"}
	}

	claudeStatus := claude.GetStatus()
	info := CycleStatusInfo{
		Type:          state.Type,
		StartedAt:     state.StartedAt,
		CurrentTaskID: state.CurrentTaskID,
		ProjectPath:   state.ProjectPath,
		ActiveWorkers: int(activeWorkers.Load()),
	}

	if claudeStatus.Used > 0 {
		info.Status = "running"
	} else {
		info.Status = "interrupted"
	}

	return info
}
