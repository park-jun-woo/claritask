# TASK-DEV-007: Memo 서비스

## 목적
Memo CRUD 비즈니스 로직 구현 (scope 기반 메모 시스템)

## 구현 파일
- `internal/service/memo_service.go` - Memo 관련 비즈니스 로직

## 상세 요구사항

### 1. Memo CRUD
```go
// SetMemo - Memo 저장 (upsert)
func SetMemo(db *db.DB, input MemoSetInput) error

type MemoSetInput struct {
    Scope    string // "project", "phase", "task"
    ScopeID  string // project_id, phase_id, task_id
    Key      string
    Value    string
    Priority int    // 1, 2, 3
    Summary  string
    Tags     []string
}

// GetMemo - Memo 조회
func GetMemo(db *db.DB, scope, scopeID, key string) (*model.Memo, error)

// DeleteMemo - Memo 삭제
func DeleteMemo(db *db.DB, scope, scopeID, key string) error
```

### 2. Memo 목록 조회
```go
// ListMemos - 전체 Memo 목록
func ListMemos(db *db.DB) (*MemoListResult, error)

// ListMemosByScope - Scope별 Memo 목록
func ListMemosByScope(db *db.DB, scope, scopeID string) ([]model.Memo, error)

type MemoListResult struct {
    Project map[string][]MemoSummary
    Phase   map[string][]MemoSummary
    Task    map[string][]MemoSummary
    Total   int
}
```

### 3. Priority 1 Memo 조회 (Manifest용)
```go
// GetHighPriorityMemos - priority=1인 memo만 조회
func GetHighPriorityMemos(db *db.DB) ([]model.Memo, error)
```

### 4. Scope 파싱
```go
// ParseMemoKey - "PH001:T042:key" 형식 파싱
func ParseMemoKey(input string) (scope, scopeID, key string, err error)

// 예시:
// "jwt_config" → project, "", "jwt_config"
// "PH001:api_decisions" → phase, "PH001", "api_decisions"
// "PH001:T042:notes" → task, "PH001:T042", "notes"
```

## 의존성
- 선행 Task: TASK-DEV-002 (database), TASK-DEV-001 (models)
- 필요 패키지: encoding/json, strings

## 완료 기준
- [ ] Memo CRUD 함수 구현
- [ ] Scope 파싱 로직 구현
- [ ] Priority 기반 필터링 구현
- [ ] JSON data 필드 처리 (value, summary, tags)
