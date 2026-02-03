# TASK-TTY-002: Phase 1 TTY Handover (init)

## 개요
`clari init`에서 Claude와 대화형으로 요구사항 수립

## 배경
- **스펙**: specs/TTY/03-Phase1.md
- **현재 상태**: headless 모드(`claude --print`)만 사용

## 작업 내용

### 1. init_service.go 수정
**파일**: `cli/internal/service/init_service.go`

```go
// RunInteractiveInit starts interactive requirements gathering
func RunInteractiveInit(database *db.DB, projectID, projectName, description string) error {
    fmt.Println("[Claritask] Starting Phase 1: Requirements Gathering")
    fmt.Printf("   Project: %s (%s)\n", projectName, projectID)
    fmt.Println("   Claude Code will help you define features.")
    fmt.Println()

    systemPrompt := Phase1SystemPrompt(projectID, projectName)
    initialPrompt := fmt.Sprintf(`프로젝트: %s
설명: %s

위 프로젝트에 필요한 기능들을 함께 정의해봅시다.`, projectName, description)

    err := RunWithTTYHandover(systemPrompt, initialPrompt, "")

    fmt.Println()
    fmt.Println("[Claritask] Phase 1 Session Ended.")

    return err
}
```

### 2. init.go 명령어 수정
**파일**: `cli/internal/cmd/init.go`

```go
var initCmd = &cobra.Command{
    Use:   "init <project_id> [description]",
    Short: "Initialize a new project",
    Args:  cobra.RangeArgs(1, 2),
    RunE:  runInit,
}

var interactiveFlag bool

func init() {
    initCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false,
        "Start interactive requirements gathering with Claude")
}

func runInit(cmd *cobra.Command, args []string) error {
    // ... 기존 초기화 로직 ...

    // 프로젝트 생성 후
    if interactiveFlag {
        return service.RunInteractiveInit(database, projectID, name, description)
    }

    // 기존 headless 분석 로직
    return nil
}
```

### 3. 플로우
```
$ clari init my-project "중고거래 플랫폼" -i

[Claritask] Starting Phase 1: Requirements Gathering
   Project: my-project
   Claude Code will help you define features.

[Claude Code 세션 시작]

사용자: "당근마켓처럼 만들어줘"
Claude: "다음 기능들을 제안합니다: ..."
...
사용자: "좋아, 개발해"
Claude: $ clari project start
[Claude Code 종료]

[Claritask] Phase 1 Session Ended.
```

## 완료 기준
- [ ] RunInteractiveInit 함수 구현
- [ ] init.go에 --interactive 플래그 추가
- [ ] Phase 1 플로우 동작 확인
