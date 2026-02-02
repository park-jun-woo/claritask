# TASK-DEV-016: Phase 명령어

## 목적
`talos phase` 서브커맨드들 구현 (create, list, plan, start)

## 구현 파일
- `internal/cmd/phase.go` - phase 명령어들

## 상세 요구사항

### 1. phase 명령어 그룹
```go
var phaseCmd = &cobra.Command{
    Use:   "phase",
    Short: "Phase management commands",
}

func init() {
    phaseCmd.AddCommand(phaseCreateCmd)
    phaseCmd.AddCommand(phaseListCmd)
}
```

### 2. phase create
```go
var phaseCreateCmd = &cobra.Command{
    Use:   "create '<json>'",
    Short: "Create a new phase",
    Args:  cobra.ExactArgs(1),
    RunE:  runPhaseCreate,
}

// JSON 입력:
// {
//   "name": "UI Planning",
//   "description": "User interface design phase",
//   "order_num": 1
// }
```

### 3. phase list
```go
var phaseListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all phases",
    RunE:  runPhaseList,
}

// 응답: phases 배열 + tasks 통계
```

### 4. phase <id> plan
```go
// 동적 서브커맨드로 처리
// talos phase 1 plan
// talos phase 2 start

func runPhasePlan(phaseID int64) error {
    // Phase 확인
    // planning 모드 안내
}
```

### 5. phase <id> start
```go
func runPhaseStart(phaseID int64) error {
    // Phase의 pending task만 실행
    // execution 모드 안내
}
```

### 6. 응답 형식
```json
// phase create
{
  "success": true,
  "phase_id": 1,
  "name": "UI Planning",
  "message": "Phase created successfully"
}

// phase list
{
  "phases": [
    {
      "id": 1,
      "name": "UI Planning",
      "order_num": 1,
      "status": "done",
      "tasks_total": 5,
      "tasks_done": 5
    }
  ],
  "total": 1
}
```

## 의존성
- 선행 Task: TASK-DEV-005 (phase-service), TASK-DEV-009 (root)
- 필요 패키지: github.com/spf13/cobra, strconv

## 완료 기준
- [ ] phase create 명령어 (JSON 파싱)
- [ ] phase list 명령어 (tasks 통계 포함)
- [ ] phase <id> plan 명령어
- [ ] phase <id> start 명령어
