# TASK-TTY-003: task run / retry 명령어

## 개요
개별 Task를 TTY Handover로 실행하는 명령어 추가

## 배경
- **스펙**: specs/TTY/05-Implementation.md
- **현재 상태**: task run, retry 명령어 없음

## 작업 내용

### 1. task.go에 명령어 추가
**파일**: `cli/internal/cmd/task.go`

```go
var taskRunCmd = &cobra.Command{
    Use:   "run <task_id>",
    Short: "Run a task with TTY handover",
    Args:  cobra.ExactArgs(1),
    RunE:  runTaskRun,
}

var taskRetryCmd = &cobra.Command{
    Use:   "retry <task_id>",
    Short: "Retry a failed task",
    Args:  cobra.ExactArgs(1),
    RunE:  runTaskRetry,
}

func init() {
    // 기존 명령어들...
    taskCmd.AddCommand(taskRunCmd)
    taskCmd.AddCommand(taskRetryCmd)
}

func runTaskRun(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        return err
    }
    defer database.Close()

    taskID, err := strconv.ParseInt(args[0], 10, 64)
    if err != nil {
        outputError(fmt.Errorf("invalid task ID: %s", args[0]))
        return nil
    }

    task, err := service.GetTask(database, taskID)
    if err != nil {
        outputError(err)
        return nil
    }

    // Start task
    if task.Status == "pending" {
        service.StartTask(database, taskID)
    }

    // Run with TTY
    if err := service.RunTaskWithTTY(database, task); err != nil {
        service.FailTask(database, taskID, err.Error())
        outputError(err)
        return nil
    }

    // Verify and complete
    if passed, err := service.VerifyTask(task); passed {
        service.CompleteTask(database, taskID, "Completed via TTY session")
        outputJSON(map[string]interface{}{
            "success": true,
            "task_id": taskID,
            "status":  "done",
            "message": "Task completed successfully",
        })
    } else {
        service.FailTask(database, taskID, err.Error())
        outputJSON(map[string]interface{}{
            "success": false,
            "task_id": taskID,
            "status":  "failed",
            "error":   err.Error(),
        })
    }

    return nil
}

func runTaskRetry(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        return err
    }
    defer database.Close()

    taskID, err := strconv.ParseInt(args[0], 10, 64)
    if err != nil {
        outputError(fmt.Errorf("invalid task ID: %s", args[0]))
        return nil
    }

    task, err := service.GetTask(database, taskID)
    if err != nil {
        outputError(err)
        return nil
    }

    if task.Status != "failed" {
        outputError(fmt.Errorf("task %d is not failed (status: %s)", taskID, task.Status))
        return nil
    }

    // Reset to pending and run
    service.ResetTaskToPending(database, taskID)

    return runTaskRun(cmd, args)
}
```

### 2. tty_service.go에 RunTaskWithTTY 추가
```go
// RunTaskWithTTY runs a single task with TTY handover
func RunTaskWithTTY(database *db.DB, task *model.Task) error {
    fmt.Printf("[Claritask] Running Task %d: %s\n", task.ID, task.Title)
    if task.TargetFile != "" {
        fmt.Printf("   Target: %s\n", task.TargetFile)
    }
    fmt.Println()

    systemPrompt := Phase2SystemPrompt()
    initialPrompt := BuildTaskPromptForTTY(database, task)

    err := RunWithTTYHandover(systemPrompt, initialPrompt, "acceptEdits")

    fmt.Println()
    fmt.Println("[Claritask] Task Session Ended.")

    return err
}
```

## 완료 기준
- [ ] taskRunCmd 명령어 구현
- [ ] taskRetryCmd 명령어 구현
- [ ] RunTaskWithTTY 함수 구현
- [ ] 사후 검증 연동
