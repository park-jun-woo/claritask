package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

const maxCycleIterations = 10

// Cycle runs full cycle: 1회차 (Plan 생성, 반복) + 2회차 (실행)
func Cycle(projectPath string) types.Result {
	ResetCancel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startTime := time.Now()
	SetCycleState(CycleState{
		Running:     true,
		Type:        "cycle",
		StartedAt:   startTime,
		ProjectPath: projectPath,
	})
	SetCycleCancel(cancel)
	defer ClearCycleState()

	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var messages []string

	// Phase 1: Plan all todo tasks (반복 순회 - subdivide로 생성된 신규 todo 포함)
	for i := 0; i < maxCycleIterations; i++ {
		if IsCancelled() || ctx.Err() != nil {
			messages = append(messages, "🛑 중단 요청으로 Plan 순회 중단")
			break
		}

		var todoCount int
		localDB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'todo'`).Scan(&todoCount)

		if todoCount == 0 {
			if i == 0 {
				messages = append(messages, "📋 Plan 순회: Plan 생성할 작업 없음")
			}
			break
		}

		messages = append(messages, fmt.Sprintf("📋 Plan 순회 %d회차: %d개 작업 Plan 생성 시작", i+1, todoCount))
		planResult := planAllInternal(ctx, projectPath)
		messages = append(messages, planResult.Message)

		if !planResult.Success {
			return types.Result{
				Success: false,
				Message: strings.Join(messages, "\n\n"),
			}
		}
	}

	// Check cancel before Phase 2
	if IsCancelled() || ctx.Err() != nil {
		messages = append(messages, "🛑 중단 요청으로 Run 순회 건너뜀")
		if globalNotifier != nil {
			notification := fmt.Sprintf("🛑 Cycle 중단됨\n소요: %s\n%s",
				formatDuration(time.Since(startTime)), strings.Join(messages, "\n"))
			globalNotifier(nil, notification)
		}
		return types.Result{
			Success: true,
			Message: strings.Join(messages, "\n\n"),
		}
	}

	// Phase 2: Run all planned tasks
	var plannedCount int
	localDB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = 'planned'`).Scan(&plannedCount)

	if plannedCount > 0 {
		messages = append(messages, fmt.Sprintf("🔄 2회차 순회: %d개 작업 실행 시작", plannedCount))
		runResult := runAllInternal(ctx, projectPath)
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

	if globalNotifier != nil {
		notification := fmt.Sprintf("🏁 Cycle 순회 완료\n소요: %s\n결과: done %d개, failed %d개",
			formatDuration(time.Since(startTime)), doneCount, failedCount)
		globalNotifier(nil, notification)
	}

	return types.Result{
		Success: true,
		Message: strings.Join(messages, "\n\n"),
	}
}
