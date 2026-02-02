# TASK-DEV-008: State 서비스

## 목적
State key-value 저장소 관리 로직 구현

## 구현 파일
- `internal/service/state_service.go` - State 관련 비즈니스 로직

## 상세 요구사항

### 1. State CRUD
```go
// SetState - State 저장 (upsert)
func SetState(db *db.DB, key, value string) error

// GetState - State 조회
func GetState(db *db.DB, key string) (string, error)

// GetAllStates - 전체 State 조회
func GetAllStates(db *db.DB) (map[string]string, error)

// DeleteState - State 삭제
func DeleteState(db *db.DB, key string) error
```

### 2. 자동 관리 State 키
```go
const (
    StateCurrentProject = "current_project"
    StateCurrentPhase   = "current_phase"
    StateCurrentTask    = "current_task"
    StateNextTask       = "next_task"
)

// UpdateCurrentState - Task 실행 시 자동 업데이트
func UpdateCurrentState(db *db.DB, projectID, phaseID string, taskID, nextTaskID int64) error
```

### 3. State 초기화
```go
// InitState - 프로젝트 초기화 시 state 설정
func InitState(db *db.DB, projectID string) error
```

## 의존성
- 선행 Task: TASK-DEV-002 (database)
- 필요 패키지: database/sql

## 완료 기준
- [ ] State CRUD 함수 구현
- [ ] 자동 관리 키 상수 정의
- [ ] Task 실행 시 state 자동 업데이트 로직
