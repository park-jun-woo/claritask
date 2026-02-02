# TASK-DEV-011: Project 명령어

## 목적
`talos project` 서브커맨드들 구현 (set, get, plan, start)

## 구현 파일
- `internal/cmd/project.go` - project 명령어들

## 상세 요구사항

### 1. project 명령어 그룹
```go
var projectCmd = &cobra.Command{
    Use:   "project",
    Short: "Project management commands",
}

func init() {
    projectCmd.AddCommand(projectSetCmd)
    projectCmd.AddCommand(projectGetCmd)
    projectCmd.AddCommand(projectPlanCmd)
    projectCmd.AddCommand(projectStartCmd)
}
```

### 2. project set
```go
var projectSetCmd = &cobra.Command{
    Use:   "set '<json>'",
    Short: "Create or update project",
    Args:  cobra.ExactArgs(1),
    RunE:  runProjectSet,
}

// JSON 입력:
// {
//   "name": "Blog Platform",
//   "description": "...",
//   "context": {...},
//   "tech": {...},
//   "design": {...}
// }
```

### 3. project get
```go
var projectGetCmd = &cobra.Command{
    Use:   "get",
    Short: "Get project information",
    RunE:  runProjectGet,
}

// 응답: project + context + tech + design 전체
```

### 4. project plan
```go
var projectPlanCmd = &cobra.Command{
    Use:   "plan",
    Short: "Start project planning",
    RunE:  runProjectPlan,
}

// 1. required 체크
// 2. planning 모드 시작 안내
```

### 5. project start
```go
var projectStartCmd = &cobra.Command{
    Use:   "start",
    Short: "Start project execution",
    RunE:  runProjectStart,
}

// 1. pending task 확인
// 2. execution 모드 시작 안내
```

## 의존성
- 선행 Task: TASK-DEV-004 (project-service), TASK-DEV-009 (root)
- 필요 패키지: github.com/spf13/cobra

## 완료 기준
- [ ] project set 명령어 (JSON 파싱, 필수 필드 검증)
- [ ] project get 명령어
- [ ] project plan 명령어 (required 체크 포함)
- [ ] project start 명령어
