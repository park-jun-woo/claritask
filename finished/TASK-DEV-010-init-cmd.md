# TASK-DEV-010: Init 명령어

## 목적
`talos init <project-id> ["<description>"]` 명령어 구현

## 구현 파일
- `internal/cmd/init.go` - init 명령어

## 상세 요구사항

### 1. 명령어 정의
```go
var initCmd = &cobra.Command{
    Use:   "init <project-id> [description]",
    Short: "Initialize a new project",
    Args:  cobra.RangeArgs(1, 2),
    RunE:  runInit,
}
```

### 2. 동작 흐름
```go
func runInit(cmd *cobra.Command, args []string) error {
    projectID := args[0]
    description := ""
    if len(args) > 1 {
        description = args[1]
    }

    // 1. projectID 유효성 검사 (영문 소문자, 숫자, -, _ 만)
    // 2. 폴더 생성 (이미 있으면 에러)
    // 3. .talos/ 디렉토리 생성
    // 4. .talos/db 파일 생성 및 스키마 초기화
    // 5. projects 테이블에 프로젝트 등록
    // 6. CLAUDE.md 템플릿 생성
    // 7. 성공 JSON 출력
}
```

### 3. 유효성 검사
```go
// validateProjectID - 프로젝트 ID 검증
func validateProjectID(id string) error {
    // 허용: a-z, 0-9, -, _
    // 불가: 대문자, 특수문자, 공백
    matched, _ := regexp.MatchString(`^[a-z0-9_-]+$`, id)
    if !matched {
        return fmt.Errorf("invalid project ID: %s", id)
    }
    return nil
}
```

### 4. CLAUDE.md 템플릿
```go
const claudeTemplate = `# %s

## Description
%s

## Tech Stack
- Backend:
- Frontend:
- Database:

## Commands
- ` + "`talos project set '<json>'`" + ` - 프로젝트 설정
- ` + "`talos required`" + ` - 필수 입력 확인
- ` + "`talos project plan`" + ` - 플래닝 시작
- ` + "`talos project start`" + ` - 실행 시작
`
```

### 5. 응답 형식
```json
{
  "success": true,
  "project_id": "blog-api",
  "path": "/path/to/blog-api",
  "message": "Project initialized successfully"
}
```

## 의존성
- 선행 Task: TASK-DEV-002 (database), TASK-DEV-009 (root)
- 필요 패키지: os, path/filepath, regexp

## 완료 기준
- [ ] init 명령어 구현
- [ ] projectID 유효성 검사
- [ ] 폴더 및 .talos/db 생성
- [ ] CLAUDE.md 템플릿 생성
- [ ] projects 테이블에 자동 등록
