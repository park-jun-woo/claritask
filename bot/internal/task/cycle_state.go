package task

import (
	"context"
	"path/filepath"
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
	ProjectID     string    `json:"project_id"`
	ActiveWorkers int       `json:"active_workers"`
	Phase         string    `json:"phase"`        // "plan", "run"
	TargetTotal   int       `json:"target_total"` // total number of tasks in current phase
	Completed     int       `json:"completed"`    // number of tasks completed in current phase
}

// activeWorkers tracks the number of currently running parallel workers (global)
var activeWorkers atomic.Int32

// Per-project cycle state management
var (
	cycleMu             sync.RWMutex
	projectCycleStates  = make(map[string]*CycleState)
	projectCycleCancels = make(map[string]context.CancelFunc)
)

// getProjectID extracts project ID from path
func getProjectID(projectPath string) string {
	return filepath.Base(projectPath)
}

// SetCycleState sets the cycle state for a specific project
func SetCycleState(projectPath string, state CycleState) {
	cycleMu.Lock()
	defer cycleMu.Unlock()
	state.ProjectID = getProjectID(projectPath)
	projectCycleStates[projectPath] = &state
}

// SetCycleCancel stores the cancel function for a specific project's cycle
func SetCycleCancel(projectPath string, cancel context.CancelFunc) {
	cycleMu.Lock()
	defer cycleMu.Unlock()
	projectCycleCancels[projectPath] = cancel
}

// CancelCycle cancels the cycle for a specific project
func CancelCycle(projectPath string) {
	cycleMu.Lock()
	cancel := projectCycleCancels[projectPath]
	cycleMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// CancelAllCycles cancels all running cycles
func CancelAllCycles() {
	cycleMu.Lock()
	cancels := make([]context.CancelFunc, 0, len(projectCycleCancels))
	for _, cancel := range projectCycleCancels {
		if cancel != nil {
			cancels = append(cancels, cancel)
		}
	}
	cycleMu.Unlock()

	for _, cancel := range cancels {
		cancel()
	}
}

// ClearCycleState clears the cycle state for a specific project
func ClearCycleState(projectPath string) {
	cycleMu.Lock()
	defer cycleMu.Unlock()
	delete(projectCycleStates, projectPath)
	delete(projectCycleCancels, projectPath)
}

// GetCycleState returns a copy of the cycle state for a specific project
func GetCycleState(projectPath string) CycleState {
	cycleMu.RLock()
	defer cycleMu.RUnlock()
	if state, ok := projectCycleStates[projectPath]; ok {
		return *state
	}
	return CycleState{}
}

// GetAllCycleStates returns all active cycle states
func GetAllCycleStates() []CycleState {
	cycleMu.RLock()
	defer cycleMu.RUnlock()
	states := make([]CycleState, 0, len(projectCycleStates))
	for _, state := range projectCycleStates {
		if state.Running {
			states = append(states, *state)
		}
	}
	return states
}

// UpdateCurrentTask updates only the CurrentTaskID field for a project
func UpdateCurrentTask(projectPath string, taskID int) {
	cycleMu.Lock()
	defer cycleMu.Unlock()
	if state, ok := projectCycleStates[projectPath]; ok {
		state.CurrentTaskID = taskID
	}
}

// UpdatePhase updates the Phase, TargetTotal and resets Completed for a project
func UpdatePhase(projectPath string, phase string, targetTotal int) {
	cycleMu.Lock()
	defer cycleMu.Unlock()
	if state, ok := projectCycleStates[projectPath]; ok {
		state.Phase = phase
		state.TargetTotal = targetTotal
		state.Completed = 0
	}
}

// IncrementCompleted increments the Completed counter by 1 for a project
func IncrementCompleted(projectPath string) {
	cycleMu.Lock()
	defer cycleMu.Unlock()
	if state, ok := projectCycleStates[projectPath]; ok {
		state.Completed++
	}
}

// UpdateActiveWorkers adjusts the active worker count by delta (+1 or -1)
func UpdateActiveWorkers(projectPath string, delta int) {
	activeWorkers.Add(int32(delta))
	cycleMu.Lock()
	if state, ok := projectCycleStates[projectPath]; ok {
		state.ActiveWorkers = int(activeWorkers.Load())
	}
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

// IsCycleRunning returns true if a cycle is running for the specified project
func IsCycleRunning(projectPath string) bool {
	cycleMu.RLock()
	defer cycleMu.RUnlock()
	if state, ok := projectCycleStates[projectPath]; ok {
		return state.Running
	}
	return false
}

// IsAnyCycleRunning returns true if any cycle is running
func IsAnyCycleRunning() bool {
	cycleMu.RLock()
	defer cycleMu.RUnlock()
	for _, state := range projectCycleStates {
		if state.Running {
			return true
		}
	}
	return false
}

// CycleStatusInfo represents the current cycle status for display
type CycleStatusInfo struct {
	Status        string    // "running", "interrupted", "idle"
	Type          string    // "cycle", "plan", "run"
	StartedAt     time.Time
	CurrentTaskID int
	ProjectPath   string
	ProjectID     string
	ActiveWorkers int
	Phase         string // "plan", "run"
	TargetTotal   int    // total number of tasks in current phase
	Completed     int    // number of tasks completed in current phase
}

// GetCycleStatus returns the combined cycle status (for backward compatibility)
// Shows the first active cycle found
func GetCycleStatus() CycleStatusInfo {
	states := GetAllCycleStates()
	if len(states) == 0 {
		return CycleStatusInfo{Status: "idle"}
	}

	// Return first active cycle for backward compatibility
	state := states[0]
	claudeStatus := claude.GetStatus()

	info := CycleStatusInfo{
		Type:          state.Type,
		StartedAt:     state.StartedAt,
		CurrentTaskID: state.CurrentTaskID,
		ProjectPath:   state.ProjectPath,
		ProjectID:     state.ProjectID,
		ActiveWorkers: int(activeWorkers.Load()),
		Phase:         state.Phase,
		TargetTotal:   state.TargetTotal,
		Completed:     state.Completed,
	}

	if claudeStatus.Used > 0 {
		info.Status = "running"
	} else {
		info.Status = "interrupted"
	}

	return info
}

// GetAllCycleStatuses returns status info for all active cycles
func GetAllCycleStatuses() []CycleStatusInfo {
	states := GetAllCycleStates()
	if len(states) == 0 {
		return nil
	}

	claudeStatus := claude.GetStatus()
	status := "running"
	if claudeStatus.Used == 0 {
		status = "interrupted"
	}

	infos := make([]CycleStatusInfo, 0, len(states))
	for _, state := range states {
		infos = append(infos, CycleStatusInfo{
			Status:        status,
			Type:          state.Type,
			StartedAt:     state.StartedAt,
			CurrentTaskID: state.CurrentTaskID,
			ProjectPath:   state.ProjectPath,
			ProjectID:     state.ProjectID,
			ActiveWorkers: state.ActiveWorkers,
			Phase:         state.Phase,
			TargetTotal:   state.TargetTotal,
			Completed:     state.Completed,
		})
	}

	return infos
}
