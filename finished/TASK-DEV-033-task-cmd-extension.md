# TASK-DEV-033: Task 커맨드 확장

## 개요
- **파일**: `internal/cmd/task.go`
- **유형**: 수정
- **우선순위**: Medium
- **Phase**: 4 (TTY Handover)
- **예상 LOC**: +100

## 목적
`clari task debug` 및 `clari task retry` 명령어 추가

## 작업 내용

### 1. task debug 명령어

```go
var taskDebugCmd = &cobra.Command{
    Use:   "debug <task_id>",
    Short: "Start interactive debugging for a task",
    Args:  cobra.ExactArgs(1),
    RunE:  runTaskDebug,
}

func runTaskDebug(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        outputError(fmt.Errorf("open database: %w", err))
        return nil
    }
    defer database.Close()

    taskID, err := strconv.ParseInt(args[0], 10, 64)
    if err != nil {
        outputError(fmt.Errorf("invalid task ID: %s", args[0]))
        return nil
    }

    task, err := service.GetTask(database, taskID)
    if err != nil {
        outputError(fmt.Errorf("get task: %w", err))
        return nil
    }

    // 대화형 디버깅 시작 (이 명령어는 블로킹)
    fmt.Println("Starting interactive debugging session...")
    fmt.Println("Press Ctrl+C to exit.")
    fmt.Println()

    if err := service.RunInteractiveDebugging(database, task); err != nil {
        outputError(fmt.Errorf("debugging failed: %w", err))
        return nil
    }

    // 검증
    passed, err := service.VerifyAfterDebugging(task)
    if !passed {
        outputJSON(map[string]interface{}{
            "success":  false,
            "task_id":  taskID,
            "verified": false,
            "error":    err.Error(),
        })
        return nil
    }

    // Task 완료 처리
    service.CompleteTask(database, taskID, "Fixed via interactive debugging")

    outputJSON(map[string]interface{}{
        "success":  true,
        "task_id":  taskID,
        "verified": true,
        "status":   "done",
        "message":  "Task completed after debugging",
    })

    return nil
}
```

### 2. task retry 명령어

```go
var taskRetryCmd = &cobra.Command{
    Use:   "retry <task_id>",
    Short: "Retry a failed task",
    Args:  cobra.ExactArgs(1),
    RunE:  runTaskRetry,
}

func init() {
    taskRetryCmd.Flags().Bool("interactive", false, "Use interactive mode for retry")
}

func runTaskRetry(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        outputError(fmt.Errorf("open database: %w", err))
        return nil
    }
    defer database.Close()

    taskID, err := strconv.ParseInt(args[0], 10, 64)
    if err != nil {
        outputError(fmt.Errorf("invalid task ID: %s", args[0]))
        return nil
    }

    interactive, _ := cmd.Flags().GetBool("interactive")

    task, err := service.GetTask(database, taskID)
    if err != nil {
        outputError(fmt.Errorf("get task: %w", err))
        return nil
    }

    // failed 상태만 retry 가능
    if task.Status != "failed" {
        outputJSON(map[string]interface{}{
            "success": false,
            "error":   fmt.Sprintf("Task status must be 'failed' to retry, current: %s", task.Status),
        })
        return nil
    }

    // Task 상태를 pending으로 재설정
    if err := service.ResetTaskToPending(database, taskID); err != nil {
        outputError(fmt.Errorf("reset task: %w", err))
        return nil
    }

    if interactive {
        // 대화형 모드로 재시도
        if err := service.RunInteractiveDebugging(database, task); err != nil {
            outputError(fmt.Errorf("interactive retry failed: %w", err))
            return nil
        }

        passed, _ := service.VerifyAfterDebugging(task)
        if passed {
            service.CompleteTask(database, taskID, "Fixed via interactive retry")
            outputJSON(map[string]interface{}{
                "success":  true,
                "task_id":  taskID,
                "status":   "done",
                "message":  "Task completed after interactive retry",
            })
        } else {
            service.FailTask(database, taskID, "Interactive retry failed")
            outputJSON(map[string]interface{}{
                "success":  false,
                "task_id":  taskID,
                "status":   "failed",
                "message":  "Task still failing after interactive retry",
            })
        }
    } else {
        // pending으로 재설정만
        outputJSON(map[string]interface{}{
            "success": true,
            "task_id": taskID,
            "status":  "pending",
            "message": "Task reset to pending. Use 'clari task pop' to execute.",
        })
    }

    return nil
}
```

### 3. ResetTaskToPending 서비스 함수 추가

`internal/service/task_service.go`에 추가:

```go
// ResetTaskToPending - Task를 pending 상태로 재설정
func ResetTaskToPending(db *db.DB, id int64) error {
    _, err := db.Exec(`
        UPDATE tasks
        SET status = 'pending',
            started_at = NULL,
            completed_at = NULL,
            failed_at = NULL,
            result = '',
            error = ''
        WHERE id = ?`, id)
    if err != nil {
        return fmt.Errorf("reset task: %w", err)
    }
    return nil
}
```

### 4. init에 커맨드 등록

```go
func init() {
    taskCmd.AddCommand(taskPushCmd)
    taskCmd.AddCommand(taskPopCmd)
    taskCmd.AddCommand(taskStartCmd)
    taskCmd.AddCommand(taskCompleteCmd)
    taskCmd.AddCommand(taskFailCmd)
    taskCmd.AddCommand(taskStatusCmd)
    taskCmd.AddCommand(taskGetCmd)
    taskCmd.AddCommand(taskListCmd)
    taskCmd.AddCommand(taskDebugCmd)  // 추가
    taskCmd.AddCommand(taskRetryCmd)  // 추가
}
```

## 의존성
- TASK-DEV-032 (Debug 서비스)

## 완료 기준
- [ ] task debug 명령어 구현됨
- [ ] task retry 명령어 구현됨
- [ ] --interactive 옵션 구현됨
- [ ] ResetTaskToPending 구현됨
- [ ] JSON 출력 형식 준수
- [ ] go build 성공
