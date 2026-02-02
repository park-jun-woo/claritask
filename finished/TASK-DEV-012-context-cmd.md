# TASK-DEV-012: Context 명령어

## 목적
`talos context set '<json>'` 명령어 구현

## 구현 파일
- `internal/cmd/context.go` - context 명령어

## 상세 요구사항

### 1. 명령어 정의
```go
var contextCmd = &cobra.Command{
    Use:   "context",
    Short: "Context management",
}

var contextSetCmd = &cobra.Command{
    Use:   "set '<json>'",
    Short: "Set project context",
    Args:  cobra.ExactArgs(1),
    RunE:  runContextSet,
}

func init() {
    contextCmd.AddCommand(contextSetCmd)
}
```

### 2. JSON 입력 형식
```json
{
  "project_name": "Blog Platform",
  "description": "Developer blogging platform",
  "target_users": "Tech bloggers",
  "deadline": "2026-03-01",
  "constraints": "Must support 10k concurrent users"
}
```

### 3. 필수 필드 검증
```go
// 필수: project_name, description
func validateContext(data map[string]interface{}) error {
    required := []string{"project_name", "description"}
    for _, field := range required {
        if _, ok := data[field]; !ok {
            return fmt.Errorf("missing required field: %s", field)
        }
    }
    return nil
}
```

### 4. 응답 형식
```json
{
  "success": true,
  "message": "Context updated successfully"
}
```

## 의존성
- 선행 Task: TASK-DEV-004 (project-service), TASK-DEV-009 (root)
- 필요 패키지: github.com/spf13/cobra, encoding/json

## 완료 기준
- [ ] context set 명령어 구현
- [ ] 필수 필드 검증 (project_name, description)
- [ ] JSON 응답 출력
