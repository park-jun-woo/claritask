# TASK-DEV-031: Project 커맨드 확장

## 개요
- **파일**: `internal/cmd/project.go`
- **유형**: 수정
- **우선순위**: High
- **Phase**: 3 (자동 실행)
- **예상 LOC**: +150

## 목적
`clari project start/stop/status` 명령어 옵션 확장

## 작업 내용

### 1. project start 옵션 추가

```go
var projectStartCmd = &cobra.Command{
    Use:   "start",
    Short: "Start project execution",
    RunE:  runProjectStart,
}

func init() {
    projectStartCmd.Flags().Int64("feature", 0, "Execute specific feature only")
    projectStartCmd.Flags().Bool("dry-run", false, "Show execution plan without running")
    projectStartCmd.Flags().Bool("fallback-interactive", false, "Switch to interactive mode on failure")
}
```

### 2. runProjectStart 확장

```go
func runProjectStart(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        outputError(fmt.Errorf("open database: %w", err))
        return nil
    }
    defer database.Close()

    // 옵션 파싱
    featureID, _ := cmd.Flags().GetInt64("feature")
    dryRun, _ := cmd.Flags().GetBool("dry-run")
    fallbackInteractive, _ := cmd.Flags().GetBool("fallback-interactive")

    // dry-run 모드
    if dryRun {
        var fid *int64
        if featureID > 0 {
            fid = &featureID
        }
        plan, err := service.GenerateExecutionPlan(database, fid)
        if err != nil {
            outputError(err)
            return nil
        }
        outputJSON(map[string]interface{}{
            "success":         true,
            "mode":            "dry-run",
            "execution_order": plan.Tasks,
            "total_tasks":     plan.Total,
        })
        return nil
    }

    // 실행 모드
    options := service.ExecutionOptions{
        DryRun:              false,
        FallbackInteractive: fallbackInteractive,
    }
    if featureID > 0 {
        options.FeatureID = &featureID
    }

    // 백그라운드 실행 시작
    go service.ExecuteAllTasks(database, options)

    outputJSON(map[string]interface{}{
        "success": true,
        "mode":    "execution",
        "message": "Execution started. Use 'clari project status' to monitor.",
    })

    return nil
}
```

### 3. project stop 명령어 추가

```go
var projectStopCmd = &cobra.Command{
    Use:   "stop",
    Short: "Stop project execution",
    RunE:  runProjectStop,
}

func runProjectStop(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        outputError(fmt.Errorf("open database: %w", err))
        return nil
    }
    defer database.Close()

    state, err := service.GetExecutionState(database)
    if err != nil {
        outputError(err)
        return nil
    }

    if !state.Running {
        outputJSON(map[string]interface{}{
            "success": false,
            "error":   "No execution in progress",
        })
        return nil
    }

    if err := service.StopExecution(database); err != nil {
        outputError(err)
        return nil
    }

    outputJSON(map[string]interface{}{
        "success":      true,
        "message":      "Execution will stop after current task completes",
        "current_task": state.CurrentTask,
    })

    return nil
}
```

### 4. project status 명령어 추가

```go
var projectStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show project execution status",
    RunE:  runProjectStatus,
}

func runProjectStatus(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        outputError(fmt.Errorf("open database: %w", err))
        return nil
    }
    defer database.Close()

    state, err := service.GetExecutionState(database)
    if err != nil {
        outputError(err)
        return nil
    }

    progress, err := service.GetExecutionProgress(database)
    if err != nil {
        outputError(err)
        return nil
    }

    outputJSON(map[string]interface{}{
        "success":   true,
        "execution": state,
        "progress":  progress,
    })

    return nil
}
```

### 5. init에 커맨드 등록

```go
func init() {
    projectCmd.AddCommand(projectSetCmd)
    projectCmd.AddCommand(projectGetCmd)
    projectCmd.AddCommand(projectPlanCmd)
    projectCmd.AddCommand(projectStartCmd)
    projectCmd.AddCommand(projectStopCmd)   // 추가
    projectCmd.AddCommand(projectStatusCmd) // 추가
}
```

## 의존성
- TASK-DEV-030 (Orchestrator 서비스)

## 완료 기준
- [ ] --feature 옵션 구현됨
- [ ] --dry-run 옵션 구현됨
- [ ] --fallback-interactive 옵션 구현됨
- [ ] project stop 명령어 구현됨
- [ ] project status 명령어 구현됨
- [ ] JSON 출력 형식 준수
- [ ] go build 성공
