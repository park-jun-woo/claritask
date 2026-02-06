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

// Stop requests cancellation of the current traversal.
// Returns a result indicating whether a traversal was running.
func Stop() (string, bool) {
	state := GetCycleState()
	if !state.Running {
		return "순회 중이 아닙니다.", false
	}

	RequestCancel()
	CancelCycle()

	typeLabel := map[string]string{"cycle": "전체순회", "plan": "플랜순회", "run": "실행순회"}[state.Type]
	if typeLabel == "" {
		typeLabel = state.Type
	}

	msg := fmt.Sprintf("🛑 %s 즉시 중단 요청됨. 실행 중인 Claude 프로세스를 종료합니다.", typeLabel)
	if state.CurrentTaskID > 0 {
		msg += fmt.Sprintf(" (현재: Task #%d)", state.CurrentTaskID)
	}

	return msg, true
}
