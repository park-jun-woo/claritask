# TASK-DEV-005: Phase 서비스

## 파일
`internal/service/phase_service.go`

## 목표
Phase CRUD 비즈니스 로직 구현

## 작업 내용

### 1. Phase CRUD
```go
// CreatePhase - Phase 생성
func CreatePhase(db *db.DB, input PhaseCreateInput) (int64, error)

type PhaseCreateInput struct {
    ProjectID   string
    Name        string
    Description string
    OrderNum    int
}

// GetPhase - Phase 조회
func GetPhase(db *db.DB, id int64) (*model.Phase, error)

// ListPhases - Phase 목록 조회 (order_num 순)
func ListPhases(db *db.DB, projectID string) ([]PhaseListItem, error)

type PhaseListItem struct {
    ID          int64
    Name        string
    Description string
    OrderNum    int
    Status      string
    TasksTotal  int
    TasksDone   int
}

// UpdatePhaseStatus - Phase 상태 변경
func UpdatePhaseStatus(db *db.DB, id int64, status string) error
```

### 2. Phase 실행 관리
```go
// StartPhase - Phase 시작 (pending → active)
func StartPhase(db *db.DB, id int64) error

// CompletePhase - Phase 완료 (active → done)
func CompletePhase(db *db.DB, id int64) error
```

## 참조
- `specs/Commands.md` - phase 명령어 섹션
- `internal/db/db.go` - DB 레이어
- `internal/model/models.go` - 모델 정의
