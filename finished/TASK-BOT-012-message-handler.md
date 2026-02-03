# TASK-BOT-012: 메시지 핸들러

## 목표
/msg 명령어 구현 (CLI message 연동)

## 파일
`bot/internal/bot/message.go`

## 작업 내용

### 메인 핸들러
```go
func (b *Bot) HandleMessage(c telebot.Context) error {
    args := strings.Fields(c.Text())
    if len(args) < 2 {
        return b.messageHelp(c)
    }

    switch args[1] {
    case "list":
        return b.messageList(c)
    case "send":
        return b.messageSendStart(c)
    case "get":
        return b.messageGet(c, args[2:])
    default:
        return b.messageHelp(c)
    }
}
```

### /msg list
```go
func (b *Bot) messageList(c telebot.Context) error {
    // 최근 메시지 목록
    // 발신/수신 구분
    // 시간 표시
}
```

### /msg send (대화형)
```go
func (b *Bot) messageSendStart(c telebot.Context) error {
    b.state.SetWaiting(c.Sender().ID, WaitingMessageTitle)
    return c.Send("메시지 제목을 입력하세요:")
}

func (b *Bot) handleMessageTitleInput(c telebot.Context, state *UserState) error {
    state.TempData["title"] = c.Text()
    b.state.SetWaiting(c.Sender().ID, WaitingMessageContent)
    return c.Send("내용을 입력하세요:")
}

func (b *Bot) handleMessageContentInput(c telebot.Context, state *UserState) error {
    state.TempData["content"] = c.Text()
    b.state.SetWaiting(c.Sender().ID, WaitingMessageRecipient)

    // Expert 선택 버튼
    experts, _ := b.svc.ListExperts(state.CurrentProject)
    // 인라인 버튼 생성
}
```

### /msg get
```go
func (b *Bot) messageGet(c telebot.Context, args []string) error {
    // 메시지 상세 내용
}
```

## 완료 조건
- [ ] HandleMessage 구현
- [ ] messageList 구현
- [ ] messageSend 대화형 구현
- [ ] messageGet 구현
