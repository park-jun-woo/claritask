# TASK-BOT-006: 상태 관리

## 목표
대화형 명령어를 위한 사용자별 상태 관리

## 파일
`bot/internal/bot/state.go`

## 작업 내용

### 상태 타입
```go
type WaitingType int

const (
    WaitingNone WaitingType = iota
    WaitingTaskTitle
    WaitingTaskDescription
    WaitingTaskExpert
    WaitingMessageTitle
    WaitingMessageContent
    WaitingMessageRecipient
    WaitingConfirmation
)
```

### UserState 구조체
```go
type UserState struct {
    CurrentProject string
    WaitingFor     WaitingType
    TempData       map[string]interface{}
    UpdatedAt      time.Time
}
```

### StateManager 구조체
```go
type StateManager struct {
    mu     sync.RWMutex
    states map[int64]*UserState
    ttl    time.Duration
}
```

### 메서드
```go
func NewStateManager(ttl time.Duration) *StateManager
func (sm *StateManager) Get(userID int64) *UserState
func (sm *StateManager) Set(userID int64, state *UserState)
func (sm *StateManager) SetWaiting(userID int64, waiting WaitingType)
func (sm *StateManager) SetTempData(userID int64, key string, value interface{})
func (sm *StateManager) GetTempData(userID int64, key string) interface{}
func (sm *StateManager) Clear(userID int64)
func (sm *StateManager) Cleanup() // 만료된 상태 정리
```

## 완료 조건
- [ ] WaitingType 정의
- [ ] UserState 구조체 정의
- [ ] StateManager 구현
- [ ] TTL 기반 자동 정리
