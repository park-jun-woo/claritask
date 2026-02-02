# TASK-DEV-014: Design 명령어

## 목적
`talos design set '<json>'` 명령어 구현

## 구현 파일
- `internal/cmd/design.go` - design 명령어

## 상세 요구사항

### 1. 명령어 정의
```go
var designCmd = &cobra.Command{
    Use:   "design",
    Short: "Design decisions management",
}

var designSetCmd = &cobra.Command{
    Use:   "set '<json>'",
    Short: "Set design decisions",
    Args:  cobra.ExactArgs(1),
    RunE:  runDesignSet,
}

func init() {
    designCmd.AddCommand(designSetCmd)
}
```

### 2. JSON 입력 형식
```json
{
  "architecture": "Microservices",
  "auth_method": "JWT with 1h expiry",
  "api_style": "RESTful",
  "db_schema_users": "id, email, password_hash, role, created_at",
  "caching_strategy": "Cache-aside pattern"
}
```

### 3. 필수 필드 검증
```go
// 필수: architecture, auth_method, api_style
func validateDesign(data map[string]interface{}) error {
    required := []string{"architecture", "auth_method", "api_style"}
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
  "message": "Design updated successfully"
}
```

## 의존성
- 선행 Task: TASK-DEV-004 (project-service), TASK-DEV-009 (root)
- 필요 패키지: github.com/spf13/cobra, encoding/json

## 완료 기준
- [ ] design set 명령어 구현
- [ ] 필수 필드 검증 (architecture, auth_method, api_style)
- [ ] JSON 응답 출력
