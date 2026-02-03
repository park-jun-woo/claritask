# TASK-DEV-114: TTY 세션 제한 구현

## 목표
Claude Code 동시 실행 세션 수 제한 (config 기반)

## 변경 파일
- `cli/internal/service/tty_service.go`

## 작업 내용

### 1. 세션 관리 변수
```go
var (
    sessionMutex    sync.Mutex
    activeSessions  int
    sessionWaiters  chan struct{}
)

func initSessionManager() {
    sessionWaiters = make(chan struct{}, 100) // 대기열 버퍼
}
```

### 2. acquireSession 함수
```go
func acquireSession(config *Config) {
    sessionMutex.Lock()

    for activeSessions >= config.TTY.MaxParallelSessions {
        sessionMutex.Unlock()
        // 대기
        <-sessionWaiters
        sessionMutex.Lock()
    }

    activeSessions++
    sessionMutex.Unlock()
}
```

### 3. releaseSession 함수
```go
func releaseSession() {
    sessionMutex.Lock()
    activeSessions--
    sessionMutex.Unlock()

    // 대기 중인 세션에 알림
    select {
    case sessionWaiters <- struct{}{}:
    default:
    }
}
```

### 4. RunWithTTYHandoverEx 수정
```go
func RunWithTTYHandoverEx(systemPrompt, initialPrompt string, permissionMode string, completeFile string) error {
    config, _ := LoadConfig()

    // 세션 획득 (대기 가능)
    fmt.Printf("[Claritask] Waiting for available session... (%d/%d)\n",
        activeSessions, config.TTY.MaxParallelSessions)
    acquireSession(config)
    defer releaseSession()

    fmt.Println("[Claritask] Session acquired. Starting Claude Code...")

    // 기존 로직...
}
```

### 5. 세션 상태 조회
```go
func GetSessionStatus() (active, max int) {
    sessionMutex.Lock()
    defer sessionMutex.Unlock()

    config, _ := LoadConfig()
    return activeSessions, config.TTY.MaxParallelSessions
}
```

## 테스트
- MAX 3 설정에서 4개 동시 실행 시 4번째가 대기하는지 확인
- 세션 완료 시 대기 중인 세션이 시작되는지 확인

## 참고
- specs/CLI/16-Config.md
- specs/VSCode/14-CLICompatibility.md
