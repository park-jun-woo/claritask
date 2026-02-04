package task

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Cycle runs full cycle: 1회차 (Plan 생성) + 2회차 (실행)
func Cycle(projectPath string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var messages []string

	// Phase 1: Plan all spec_ready tasks
	var specReadyCount int
	localDB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'spec_ready'`).Scan(&specReadyCount)

	if specReadyCount > 0 {
		messages = append(messages, fmt.Sprintf("📋 1회차 순회: %d개 작업 Plan 생성 시작", specReadyCount))
		planResult := PlanAll(projectPath)
		messages = append(messages, planResult.Message)

		if !planResult.Success {
			return types.Result{
				Success: false,
				Message: strings.Join(messages, "\n\n"),
			}
		}
	} else {
		messages = append(messages, "📋 1회차 순회: Plan 생성할 작업 없음")
	}

	// Phase 2: Run all plan_ready tasks
	var planReadyCount int
	localDB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'plan_ready'`).Scan(&planReadyCount)

	if planReadyCount > 0 {
		messages = append(messages, fmt.Sprintf("🔄 2회차 순회: %d개 작업 실행 시작", planReadyCount))
		runResult := RunAll(projectPath)
		messages = append(messages, runResult.Message)

		if !runResult.Success {
			return types.Result{
				Success: false,
				Message: strings.Join(messages, "\n\n"),
			}
		}
	} else {
		messages = append(messages, "🔄 2회차 순회: 실행할 작업 없음")
	}

	// Summary
	var doneCount, failedCount int
	localDB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'done'`).Scan(&doneCount)
	localDB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'failed'`).Scan(&failedCount)

	messages = append(messages, fmt.Sprintf("🏁 Cycle 완료: done %d개, failed %d개", doneCount, failedCount))

	return types.Result{
		Success: true,
		Message: strings.Join(messages, "\n\n"),
	}
}
