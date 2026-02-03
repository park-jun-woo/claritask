# TASK-TTY-004: Phase 2 Orchestrator TTY 연결

## 개요
orchestrator_service.go에서 FallbackInteractive 옵션을 RunTaskWithTTY에 연결

## 배경
- **스펙**: specs/TTY/04-Phase2.md
- **현재 상태**: TODO 주석으로 미연결

## 작업 내용

### 1. orchestrator_service.go 수정
**파일**: `cli/internal/service/orchestrator_service.go`

```go
// ExecuteAllTasks executes all pending tasks
func ExecuteAllTasks(database *db.DB, options ExecutionOptions) error {
    if err := StartExecution(database); err != nil {
        return err
    }
    defer StopExecution(database)

    for {
        // Check for stop request
        if IsStopRequested(database) {
            fmt.Println("[Claritask] Stop requested. Stopping after current task.")
            return nil
        }

        // Get next executable task
        response, err := PopTaskFull(database)
        if err != nil {
            return fmt.Errorf("pop task: %w", err)
        }
        if response.Task == nil {
            fmt.Println("[Claritask] All tasks completed!")
            return nil
        }

        task := response.Task

        // Filter by feature if specified
        if options.FeatureID != nil && task.FeatureID != *options.FeatureID {
            ResetTaskToPending(database, task.ID)
            continue
        }

        // Update current task
        UpdateExecutionCurrentTask(database, task.ID)

        fmt.Printf("\n[Claritask] Executing Task %d/%d: %s\n",
            response.Manifest.CompletedTasks+1,
            response.Manifest.TotalTasks,
            task.Title)

        // Try headless execution first
        result, err := ExecuteTaskWithClaude(task, &response.Manifest)

        if err == nil && result.Success {
            CompleteTask(database, task.ID, result.Output)
            fmt.Printf("[Claritask] Task %d completed.\n", task.ID)
            continue
        }

        // Headless failed
        fmt.Printf("[Claritask] Headless execution failed: %s\n", result.Error)

        if options.FallbackInteractive {
            fmt.Println("[Claritask] Switching to interactive mode...")

            // Run with TTY Handover
            if err := RunTaskWithTTY(database, task); err != nil {
                FailTask(database, task.ID, err.Error())
                continue
            }

            // Verify after TTY session
            passed, verifyErr := VerifyTask(task)
            if passed {
                CompleteTask(database, task.ID, "Completed via interactive session")
                fmt.Printf("[Claritask] Task %d completed (interactive).\n", task.ID)
            } else {
                FailTask(database, task.ID, verifyErr.Error())
                fmt.Printf("[Claritask] Task %d failed after interactive session.\n", task.ID)
            }
        } else {
            FailTask(database, task.ID, result.Error)
            return fmt.Errorf("task %d failed: %s", task.ID, result.Error)
        }
    }
}
```

### 2. VerifyTask 함수 추가
```go
// VerifyTask verifies task completion by running test command
func VerifyTask(task *model.Task) (bool, error) {
    testCmd := InferTestCommand(task.TargetFile)
    if testCmd == "" || testCmd == "# Run appropriate test command" {
        // Cannot auto-verify
        return true, nil
    }

    fmt.Printf("[Claritask] Verifying: %s\n", testCmd)

    cmd := exec.Command("sh", "-c", testCmd)
    output, err := cmd.CombinedOutput()

    if err == nil {
        fmt.Println("[Claritask] Verification passed!")
        return true, nil
    }

    return false, fmt.Errorf("verification failed: %s", string(output))
}
```

### 3. project start 수정 (foreground 실행)
```go
// project.go
func runProjectStart(cmd *cobra.Command, args []string) error {
    // ... 기존 검증 로직 ...

    // Execute (foreground, blocking)
    fmt.Println("[Claritask] Starting Phase 2: Task Execution")
    err := service.ExecuteAllTasks(database, options)

    if err != nil {
        outputJSON(map[string]interface{}{
            "success": false,
            "error":   err.Error(),
        })
        return nil
    }

    // Final report
    progress, _ := service.GetExecutionProgress(database)
    outputJSON(map[string]interface{}{
        "success":  true,
        "mode":     "completed",
        "progress": progress,
        "message":  "Execution completed",
    })

    return nil
}
```

## 완료 기준
- [ ] FallbackInteractive → RunTaskWithTTY 연결
- [ ] VerifyTask 함수 구현
- [ ] project start를 foreground 실행으로 변경
- [ ] 최종 보고 출력
