# Claribot Architecture

> **현재 버전**: v0.0.1

---

## 시스템 구성

### 컴포넌트 다이어그램

```
┌─────────────────────────────────────────────────────────────┐
│                        Claribot                              │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Router    │  │  Handlers   │  │ Middleware  │         │
│  │             │──▶│             │──▶│  (Auth)    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│         │                │                                   │
│         ▼                ▼                                   │
│  ┌─────────────────────────────────────────────┐           │
│  │              Service Layer                   │           │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐       │           │
│  │  │ Project │ │  Task   │ │ Message │       │           │
│  │  │ Service │ │ Service │ │ Service │       │           │
│  │  └─────────┘ └─────────┘ └─────────┘       │           │
│  └─────────────────────────────────────────────┘           │
│                          │                                   │
│                          ▼                                   │
│  ┌─────────────────────────────────────────────┐           │
│  │              Data Layer                      │           │
│  │  ┌─────────────────┐  ┌─────────────────┐  │           │
│  │  │  DB (읽기/쓰기) │  │  CLI Bridge     │  │           │
│  │  │  직접 접근      │  │  (선택적)       │  │           │
│  │  └─────────────────┘  └─────────────────┘  │           │
│  └─────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘
         │                           │
         ▼                           ▼
┌─────────────────┐         ┌─────────────────┐
│ ~/.claritask/   │         │   clari CLI     │
│   db.clt        │         │   (외부 호출)   │
└─────────────────┘         └─────────────────┘
```

---

## 계층별 책임

### 1. Router Layer

텔레그램 메시지를 적절한 핸들러로 라우팅

```go
// internal/bot/router.go
func SetupRoutes(b *telebot.Bot, h *Handlers) {
    // 명령어 라우팅
    b.Handle("/start", h.Start)
    b.Handle("/project", h.Project)
    b.Handle("/task", h.Task)
    b.Handle("/msg", h.Message)

    // 콜백 쿼리 (인라인 버튼)
    b.Handle(telebot.OnCallback, h.Callback)

    // 일반 텍스트 (대화형)
    b.Handle(telebot.OnText, h.Text)
}
```

### 2. Handler Layer

각 명령어의 비즈니스 로직 처리

```go
// internal/bot/handlers.go
type Handlers struct {
    projectSvc *service.ProjectService
    taskSvc    *service.TaskService
    msgSvc     *service.MessageService
    config     *config.Config
}

func (h *Handlers) Task(c telebot.Context) error {
    // 1. 인자 파싱
    // 2. 서비스 호출
    // 3. 응답 포맷팅
    // 4. 텔레그램 응답
}
```

### 3. Middleware Layer

인증 및 권한 검사

```go
// internal/bot/middleware.go
func AuthMiddleware(allowedUsers []int64) telebot.MiddlewareFunc {
    return func(next telebot.HandlerFunc) telebot.HandlerFunc {
        return func(c telebot.Context) error {
            userID := c.Sender().ID
            if !isAllowed(userID, allowedUsers) {
                return c.Send("권한이 없습니다.")
            }
            return next(c)
        }
    }
}
```

### 4. Service Layer

CLI의 service 패키지 재사용 또는 별도 구현

```go
// internal/service/project_service.go
type ProjectService struct {
    db *db.DB
}

func (s *ProjectService) GetCurrent() (*model.Project, error) {
    // DB에서 현재 프로젝트 조회
}

func (s *ProjectService) GetStatus() (*ProjectStatus, error) {
    // 프로젝트 상태 집계
}
```

### 5. Data Layer

#### 옵션 A: DB 직접 접근 (권장)

```go
// internal/db/db.go
// CLI의 db 패키지를 import하거나 동일하게 구현
import "claritask/cli/internal/db"

func Open(path string) (*db.DB, error) {
    return db.Open(path)
}
```

#### 옵션 B: CLI 브릿지

```go
// internal/bridge/cli_bridge.go
func ExecCLI(args ...string) ([]byte, error) {
    cmd := exec.Command("clari", args...)
    return cmd.Output()
}

func GetTasks(projectID string) ([]model.Task, error) {
    out, err := ExecCLI("task", "list", "--project", projectID)
    if err != nil {
        return nil, err
    }
    var resp struct {
        Success bool         `json:"success"`
        Tasks   []model.Task `json:"tasks"`
    }
    json.Unmarshal(out, &resp)
    return resp.Tasks, nil
}
```

---

## 데이터 흐름

### 읽기 작업 (조회)

```
User ──▶ Telegram ──▶ Claribot ──▶ DB
                                    │
User ◀── Telegram ◀── Claribot ◀───┘
```

### 쓰기 작업 (생성/수정)

```
User ──▶ Telegram ──▶ Claribot ──▶ DB (직접)
                          │
                          └──▶ CLI (옵션: 복잡한 로직)
```

---

## 상태 관리

### 대화 상태 (Conversation State)

```go
// internal/bot/state.go
type ConversationState struct {
    mu     sync.RWMutex
    states map[int64]*UserState  // userID -> state
}

type UserState struct {
    CurrentProject string
    WaitingFor     WaitingType  // 입력 대기 상태
    TempData       interface{}  // 임시 데이터
    UpdatedAt      time.Time
}

type WaitingType int
const (
    WaitingNone WaitingType = iota
    WaitingTaskTitle
    WaitingMessageContent
    WaitingConfirmation
)
```

### 상태 전이 예시

```
/task add
    │
    ▼
WaitingTaskTitle ──▶ "로그인 기능 구현"
    │
    ▼
WaitingConfirmation ──▶ "Yes"
    │
    ▼
태스크 생성 완료
```

---

## 동시성 처리

### DB 접근

```go
// 읽기: 동시 접근 허용
// 쓰기: 단일 Writer (SQLite WAL 모드)

func (s *Service) WithTransaction(fn func(tx *sql.Tx) error) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if err := fn(tx); err != nil {
        return err
    }
    return tx.Commit()
}
```

### Rate Limiting

```go
// 텔레그램 API 제한 준수
// - 같은 채팅: 1초에 1개 메시지
// - 전체: 초당 30개 메시지

type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Every(time.Second/30), 1),
    }
}
```

---

## 에러 처리

```go
// internal/bot/errors.go
type BotError struct {
    Code    string
    Message string
    UserMsg string  // 사용자에게 보여줄 메시지
}

var (
    ErrUnauthorized = &BotError{
        Code:    "UNAUTHORIZED",
        Message: "unauthorized user",
        UserMsg: "권한이 없습니다.",
    }
    ErrProjectNotFound = &BotError{
        Code:    "PROJECT_NOT_FOUND",
        Message: "project not found",
        UserMsg: "프로젝트를 찾을 수 없습니다.",
    }
)
```

---

## 설정 구조

```go
// internal/config/config.go
type Config struct {
    // 텔레그램 설정
    TelegramToken string   `env:"TELEGRAM_TOKEN,required"`
    AllowedUsers  []int64  `env:"ALLOWED_USERS"`  // 허용된 사용자 ID
    AdminUsers    []int64  `env:"ADMIN_USERS"`    // 관리자 ID

    // DB 설정
    DBPath string `env:"CLARITASK_DB" envDefault:"~/.claritask/db.clt"`

    // 알림 설정
    NotifyOnComplete bool `env:"NOTIFY_ON_COMPLETE" envDefault:"true"`
    NotifyOnFail     bool `env:"NOTIFY_ON_FAIL" envDefault:"true"`

    // 로깅
    LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}
```

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [01-Overview.md](01-Overview.md) | 전체 개요 |
| [03-Commands.md](03-Commands.md) | 봇 명령어 |
| [04-Deployment.md](04-Deployment.md) | 배포 |

---

*Claribot Architecture v0.0.1*
