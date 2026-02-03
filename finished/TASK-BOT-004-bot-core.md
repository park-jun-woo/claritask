# TASK-BOT-004: Bot 코어

## 목표
Bot 구조체 및 초기화 로직 구현

## 파일
`bot/internal/bot/bot.go`

## 작업 내용

### Bot 구조체
```go
type Bot struct {
    tg      *telebot.Bot
    cfg     *config.Config
    svc     *service.Service
    state   *StateManager
    limiter *RateLimiter
    logger  zerolog.Logger
}
```

### 생성자
```go
func New(cfg *config.Config) (*Bot, error) {
    // 1. 텔레그램 봇 생성
    // 2. 서비스 초기화
    // 3. 상태 관리자 초기화
    // 4. Rate Limiter 초기화
    // 5. 라우터 설정
    // 6. 미들웨어 등록
}
```

### 메서드
```go
func (b *Bot) Start() // 봇 시작 (blocking)
func (b *Bot) Stop()  // 봇 중지
```

### 텔레그램 봇 설정
```go
pref := telebot.Settings{
    Token:  cfg.TelegramToken,
    Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
}
```

## 완료 조건
- [ ] Bot 구조체 정의
- [ ] New 생성자 구현
- [ ] Start/Stop 메서드 구현
