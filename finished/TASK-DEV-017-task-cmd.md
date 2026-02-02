# TASK-DEV-017: Task 명령어

## 목적
`talos task` 서브커맨드들 구현 (push, pop, start, complete, fail, status)

## 구현 파일
- `internal/cmd/task.go` - task 명령어들

## 상세 요구사항

### 1. task 명령어 그룹
```go
var taskCmd = &cobra.Command{
    Use:   "task",
    Short: "Task management commands",
}

func init() {
    taskCmd.AddCommand(taskPushCmd)
    taskCmd.AddCommand(taskPopCmd)
    taskCmd.AddCommand(taskStartCmd)
    taskCmd.AddCommand(taskCompleteCmd)
    taskCmd.AddCommand(taskFailCmd)
    taskCmd.AddCommand(taskStatusCmd)
}
```

### 2. task push
```go
var taskPushCmd = &cobra.Command{
    Use:   "push '<json>'",
    Short: "Add a new task",
    Args:  cobra.ExactArgs(1),
    RunE:  runTaskPush,
}

// JSON 입력:
// {
//   "phase_id": 1,
//   "title": "Setup project",
//   "content": "Create initial structure",
//   "level": "node",
//   "references": ["specs/requirements.md"]
// }
```

### 3. task pop
```go
var taskPopCmd = &cobra.Command{
    Use:   "pop",
    Short: "Get next pending task with manifest",
    RunE:  runTaskPop,
}

// 응답: task + manifest (context, tech, design, state, memos)
```

### 4. task start
```go
var taskStartCmd = &cobra.Command{
    Use:   "start <task_id>",
    Short: "Start a task (pending → doing)",
    Args:  cobra.ExactArgs(1),
    RunE:  runTaskStart,
}
```

### 5. task complete
```go
var taskCompleteCmd = &cobra.Command{
    Use:   "complete <task_id> '<json>'",
    Short: "Complete a task (doing → done)",
    Args:  cobra.ExactArgs(2),
    RunE:  runTaskComplete,
}

// JSON 입력:
// {
//   "result": "success",
//   "notes": "Completed successfully"
// }
```

### 6. task fail
```go
var taskFailCmd = &cobra.Command{
    Use:   "fail <task_id> '<json>'",
    Short: "Fail a task (doing → failed)",
    Args:  cobra.ExactArgs(2),
    RunE:  runTaskFail,
}

// JSON 입력:
// {
//   "error": "Database connection failed",
//   "details": "Connection timeout"
// }
```

### 7. task status
```go
var taskStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show task progress",
    RunE:  runTaskStatus,
}

// 응답: summary + current_phase + current_task + failed_tasks
```

## 의존성
- 선행 Task: TASK-DEV-006 (task-service), TASK-DEV-009 (root)
- 필요 패키지: github.com/spf13/cobra, strconv

## 완료 기준
- [ ] task push 명령어 (필수 필드: phase_id, title, content)
- [ ] task pop 명령어 (manifest 포함)
- [ ] task start 명령어 (상태 검증)
- [ ] task complete 명령어 (result 필수)
- [ ] task fail 명령어 (error 필수)
- [ ] task status 명령어 (진행률 계산)
