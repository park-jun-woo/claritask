# TASK-BOT-009: 기본 핸들러

## 목표
/start, /help, /status 및 텍스트/콜백 핸들러 구현

## 파일
`bot/internal/bot/handlers.go`

## 작업 내용

### /start 핸들러
```go
func (b *Bot) HandleStart(c telebot.Context) error {
    // 현재 프로젝트 정보
    // 진행률 표시
    // 환영 메시지
}
```

### /help 핸들러
```go
func (b *Bot) HandleHelp(c telebot.Context) error {
    // 명령어 도움말 표시
    // 인자가 있으면 해당 명령어 상세 도움말
}
```

### /status 핸들러
```go
func (b *Bot) HandleStatus(c telebot.Context) error {
    // 전체 대시보드
    // 프로젝트, 태스크, 메시지 요약
}
```

### 텍스트 핸들러 (대화형 입력)
```go
func (b *Bot) HandleText(c telebot.Context) error {
    userID := c.Sender().ID
    state := b.state.Get(userID)

    switch state.WaitingFor {
    case WaitingTaskTitle:
        return b.handleTaskTitleInput(c, state)
    case WaitingMessageContent:
        return b.handleMessageContentInput(c, state)
    // ... 기타 상태
    default:
        return c.Send("❓ 알 수 없는 입력입니다. /help를 확인하세요.")
    }
}
```

### 콜백 핸들러 (인라인 버튼)
```go
func (b *Bot) HandleCallback(c telebot.Context) error {
    data := c.Callback().Data
    parts := strings.Split(data, ":")

    switch parts[0] {
    case "task":
        return b.handleTaskCallback(c, parts)
    case "project":
        return b.handleProjectCallback(c, parts)
    // ... 기타
    }
}
```

### 유틸리티
```go
func progressBar(percent float64, width int) string
func formatTime(t time.Time) string
```

## 완료 조건
- [ ] HandleStart 구현
- [ ] HandleHelp 구현
- [ ] HandleStatus 구현
- [ ] HandleText 구현
- [ ] HandleCallback 구현
