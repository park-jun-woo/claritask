# TASK-DEV-006: Task 서비스

## 목적
Task CRUD 및 상태 전이 비즈니스 로직 구현

## 구현 파일
- `internal/service/task_service.go` - Task 관련 비즈니스 로직

## 상세 요구사항

### 1. Task CRUD
```go
// CreateTask - Task 생성
func CreateTask(db *db.DB, input TaskCreateInput) (int64, error)

type TaskCreateInput struct {
    PhaseID    int64
    ParentID   *int64
    Title      string
    Content    string
    Level      string   // "", "node", "leaf"
    Skill      string
    References []string
}

// GetTask - Task 조회
func GetTask(db *db.DB, id int64) (*model.Task, error)

// ListTasks - Phase별 Task 목록
func ListTasks(db *db.DB, phaseID int64) ([]model.Task, error)
```

### 2. Task 상태 전이
```go
// StartTask - pending → doing
func StartTask(db *db.DB, id int64) error

// CompleteTask - doing → done
func CompleteTask(db *db.DB, id int64, result string) error

// FailTask - doing → failed
func FailTask(db *db.DB, id int64, errMsg string) error
```

### 3. Task Pop (Manifest 포함)
```go
// PopTask - 다음 pending Task + Manifest 반환
func PopTask(db *db.DB) (*TaskPopResult, error)

type TaskPopResult struct {
    Task     *model.Task
    Manifest *Manifest
}

// Manifest 구성:
// - context, tech, design 전체 데이터
// - state 현재 상태
// - priority=1인 memos만
```

### 4. Task 상태 요약
```go
// GetTaskStatus - 전체 진행 상황
func GetTaskStatus(db *db.DB) (*TaskStatusResult, error)

type TaskStatusResult struct {
    Total    int
    Pending  int
    Doing    int
    Done     int
    Failed   int
    Progress float64
}
```

## 의존성
- 선행 Task: TASK-DEV-002 (database), TASK-DEV-001 (models)
- 필요 패키지: database/sql, encoding/json

## 완료 기준
- [ ] Task CRUD 함수 구현
- [ ] 상태 전이 함수 구현 (pending → doing → done/failed)
- [ ] PopTask에서 Manifest 포함하여 반환
- [ ] GetTaskStatus로 진행률 조회 가능
