# TASK-DEV-013: Tech 명령어

## 목적
`talos tech set '<json>'` 명령어 구현

## 구현 파일
- `internal/cmd/tech.go` - tech 명령어

## 상세 요구사항

### 1. 명령어 정의
```go
var techCmd = &cobra.Command{
    Use:   "tech",
    Short: "Tech stack management",
}

var techSetCmd = &cobra.Command{
    Use:   "set '<json>'",
    Short: "Set tech stack",
    Args:  cobra.ExactArgs(1),
    RunE:  runTechSet,
}

func init() {
    techCmd.AddCommand(techSetCmd)
}
```

### 2. JSON 입력 형식
```json
{
  "backend": "FastAPI",
  "frontend": "React 18",
  "database": "PostgreSQL",
  "cache": "Redis",
  "deployment": "Docker + AWS ECS"
}
```

### 3. 필수 필드 검증
```go
// 필수: backend, frontend, database
func validateTech(data map[string]interface{}) error {
    required := []string{"backend", "frontend", "database"}
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
  "message": "Tech updated successfully"
}
```

## 의존성
- 선행 Task: TASK-DEV-004 (project-service), TASK-DEV-009 (root)
- 필요 패키지: github.com/spf13/cobra, encoding/json

## 완료 기준
- [ ] tech set 명령어 구현
- [ ] 필수 필드 검증 (backend, frontend, database)
- [ ] JSON 응답 출력
