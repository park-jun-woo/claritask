# TASK-DEV-004: 프로젝트 서비스

## 파일
`internal/service/project_service.go`

## 목표
프로젝트, Context, Tech, Design CRUD 비즈니스 로직 구현

## 작업 내용

### 1. 프로젝트 관리
```go
// CreateProject - 프로젝트 생성
func CreateProject(db *db.DB, id, name, description string) error

// GetProject - 프로젝트 조회
func GetProject(db *db.DB) (*model.Project, error)

// UpdateProject - 프로젝트 업데이트
func UpdateProject(db *db.DB, p *model.Project) error
```

### 2. Context 관리
```go
// SetContext - Context 설정 (upsert)
func SetContext(db *db.DB, data map[string]interface{}) error

// GetContext - Context 조회
func GetContext(db *db.DB) (map[string]interface{}, error)
```

### 3. Tech 관리
```go
// SetTech - Tech 설정 (upsert)
func SetTech(db *db.DB, data map[string]interface{}) error

// GetTech - Tech 조회
func GetTech(db *db.DB) (map[string]interface{}, error)
```

### 4. Design 관리
```go
// SetDesign - Design 설정 (upsert)
func SetDesign(db *db.DB, data map[string]interface{}) error

// GetDesign - Design 조회
func GetDesign(db *db.DB) (map[string]interface{}, error)
```

### 5. Required 검사
```go
// CheckRequired - 필수 항목 검사
func CheckRequired(db *db.DB) (*RequiredResult, error)

type RequiredResult struct {
    Ready          bool
    MissingRequired []MissingField
}

type MissingField struct {
    Field   string
    Prompt  string
    Options []string
}
```

### 6. 전체 프로젝트 설정
```go
// SetProjectFull - 프로젝트 전체 설정 (project + context + tech + design)
func SetProjectFull(db *db.DB, input ProjectSetInput) error

type ProjectSetInput struct {
    Name        string
    Description string
    Context     map[string]interface{}
    Tech        map[string]interface{}
    Design      map[string]interface{}
}
```

## 필수 필드 목록
- context: project_name, description
- tech: backend, frontend, database
- design: architecture, auth_method, api_style

## 참조
- `specs/Commands.md` - project, context, tech, design, required 명령어
- `internal/db/db.go` - DB 레이어
- `internal/model/models.go` - 모델 정의
