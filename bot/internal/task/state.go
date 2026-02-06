package task

import (
	"fmt"
	"sync/atomic"
)

// cancelRequested is the atomic flag for graceful stop
var cancelRequested atomic.Bool

// RequestCancel sets the cancel flag to true
func RequestCancel() {
	cancelRequested.Store(true)
}

// IsCancelled returns true if cancel has been requested
func IsCancelled() bool {
	return cancelRequested.Load()
}

// ResetCancel clears the cancel flag
func ResetCancel() {
	cancelRequested.Store(false)
}

// Stop requests cancellation of all running traversals.
// Returns a result indicating whether any traversal was running.
func Stop() (string, bool) {
	states := GetAllCycleStates()
	if len(states) == 0 {
		return "순회 중인 프로젝트가 없습니다.", false
	}

	RequestCancel()
	CancelAllCycles()

	var msg string
	if len(states) == 1 {
		state := states[0]
		typeLabel := map[string]string{"cycle": "전체순회", "plan": "플랜순회", "run": "실행순회"}[state.Type]
		if typeLabel == "" {
			typeLabel = state.Type
		}
		msg = fmt.Sprintf("🛑 [%s] %s 즉시 중단 요청됨", state.ProjectID, typeLabel)
		if state.CurrentTaskID > 0 {
			msg += fmt.Sprintf(" (Task #%d)", state.CurrentTaskID)
		}
	} else {
		msg = fmt.Sprintf("🛑 %d개 프로젝트 순회 즉시 중단 요청됨:", len(states))
		for _, state := range states {
			msg += fmt.Sprintf("\n   - [%s]", state.ProjectID)
		}
	}

	return msg, true
}

// StopProject requests cancellation of a specific project's traversal.
func StopProject(projectPath string) (string, bool) {
	if !IsCycleRunning(projectPath) {
		return fmt.Sprintf("이 프로젝트는 순회 중이 아닙니다: %s", getProjectID(projectPath)), false
	}

	CancelCycle(projectPath)

	state := GetCycleState(projectPath)
	typeLabel := map[string]string{"cycle": "전체순회", "plan": "플랜순회", "run": "실행순회"}[state.Type]
	if typeLabel == "" {
		typeLabel = state.Type
	}

	msg := fmt.Sprintf("🛑 [%s] %s 즉시 중단 요청됨", getProjectID(projectPath), typeLabel)
	if state.CurrentTaskID > 0 {
		msg += fmt.Sprintf(" (Task #%d)", state.CurrentTaskID)
	}

	return msg, true
}
