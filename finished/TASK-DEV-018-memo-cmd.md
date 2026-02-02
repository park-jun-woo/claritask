# TASK-DEV-018: Memo 명령어

## 목적
`talos memo` 서브커맨드들 구현 (set, get, list, del)

## 구현 파일
- `internal/cmd/memo.go` - memo 명령어들

## 상세 요구사항

### 1. memo 명령어 그룹
```go
var memoCmd = &cobra.Command{
    Use:   "memo",
    Short: "Memo management commands",
}

func init() {
    memoCmd.AddCommand(memoSetCmd)
    memoCmd.AddCommand(memoGetCmd)
    memoCmd.AddCommand(memoListCmd)
    memoCmd.AddCommand(memoDelCmd)
}
```

### 2. memo set
```go
var memoSetCmd = &cobra.Command{
    Use:   "set <key> '<json>'",
    Short: "Set a memo",
    Args:  cobra.ExactArgs(2),
    RunE:  runMemoSet,
}

// key 형식:
// "jwt_config" → project scope
// "1:api_decisions" → phase scope (phase_id:key)
// "1:42:notes" → task scope (phase_id:task_id:key)

// JSON 입력:
// {
//   "value": "Use httpOnly cookies",
//   "priority": 1,
//   "summary": "Security best practice",
//   "tags": ["security", "jwt"]
// }
```

### 3. memo get
```go
var memoGetCmd = &cobra.Command{
    Use:   "get <key>",
    Short: "Get a memo",
    Args:  cobra.ExactArgs(1),
    RunE:  runMemoGet,
}

// key 형식 동일
```

### 4. memo list
```go
var memoListCmd = &cobra.Command{
    Use:   "list [scope]",
    Short: "List memos",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runMemoList,
}

// scope 형식:
// (없음) → 전체
// "1" → phase 1의 메모
// "1:42" → phase 1, task 42의 메모
```

### 5. memo del
```go
var memoDelCmd = &cobra.Command{
    Use:   "del <key>",
    Short: "Delete a memo",
    Args:  cobra.ExactArgs(1),
    RunE:  runMemoDel,
}
```

### 6. Key 파싱 로직
```go
// parseKey - "1:42:notes" → (scope, scopeID, key)
func parseKey(input string) (scope, scopeID, key string) {
    parts := strings.Split(input, ":")
    switch len(parts) {
    case 1:
        return "project", "", parts[0]
    case 2:
        return "phase", parts[0], parts[1]
    case 3:
        return "task", parts[0]+":"+parts[1], parts[2]
    }
    return "", "", ""
}
```

## 의존성
- 선행 Task: TASK-DEV-007 (memo-service), TASK-DEV-009 (root)
- 필요 패키지: github.com/spf13/cobra, strings

## 완료 기준
- [ ] memo set 명령어 (scope 파싱, value 필수)
- [ ] memo get 명령어
- [ ] memo list 명령어 (scope 필터링)
- [ ] memo del 명령어
- [ ] key 파싱 로직 (project/phase/task scope)
