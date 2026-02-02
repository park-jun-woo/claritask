# TASK-DEV-015: Required 명령어

## 목적
`talos required` 명령어 구현 - 필수 입력 누락 항목 확인

## 구현 파일
- `internal/cmd/required.go` - required 명령어

## 상세 요구사항

### 1. 명령어 정의
```go
var requiredCmd = &cobra.Command{
    Use:   "required",
    Short: "Check required configuration",
    RunE:  runRequired,
}
```

### 2. 필수 항목 목록
```go
var requiredFields = []RequiredField{
    // Context
    {Section: "context", Field: "project_name", Prompt: "What is the project name?", Examples: []string{"Blog Platform", "E-commerce API"}},
    {Section: "context", Field: "description", Prompt: "What is the project description?", Examples: []string{"Developer blogging platform"}},

    // Tech
    {Section: "tech", Field: "backend", Prompt: "What backend framework?", Options: []string{"FastAPI", "Django", "Flask", "Express", "Go/Gin"}},
    {Section: "tech", Field: "frontend", Prompt: "What frontend framework?", Options: []string{"React", "Vue", "Angular", "Next.js", "None"}},
    {Section: "tech", Field: "database", Prompt: "What database?", Options: []string{"PostgreSQL", "MySQL", "MongoDB", "SQLite"}},

    // Design
    {Section: "design", Field: "architecture", Prompt: "What architecture pattern?", Options: []string{"Monolithic", "Microservices", "Serverless"}},
    {Section: "design", Field: "auth_method", Prompt: "What authentication method?", Options: []string{"JWT", "Session", "OAuth", "None"}},
    {Section: "design", Field: "api_style", Prompt: "What API style?", Options: []string{"RESTful", "GraphQL", "gRPC"}},
}
```

### 3. 응답 형식 (누락 시)
```json
{
  "ready": false,
  "missing_required": [
    {
      "field": "context.project_name",
      "prompt": "What is the project name?",
      "examples": ["Blog Platform", "E-commerce API"]
    },
    {
      "field": "tech.backend",
      "prompt": "What backend framework?",
      "options": ["FastAPI", "Django", "Flask", "Express"],
      "custom_allowed": true
    }
  ],
  "total_missing": 2,
  "message": "Please configure required settings"
}
```

### 4. 응답 형식 (완료 시)
```json
{
  "ready": true,
  "message": "All required fields configured"
}
```

## 의존성
- 선행 Task: TASK-DEV-004 (project-service), TASK-DEV-009 (root)
- 필요 패키지: github.com/spf13/cobra

## 완료 기준
- [ ] required 명령어 구현
- [ ] context, tech, design 각 섹션 검사
- [ ] 누락 항목에 대한 안내 메시지 포함
- [ ] ready 상태 반환
