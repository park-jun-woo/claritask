# TASK-BOT-013: Expert 핸들러

## 목표
/expert 명령어 구현

## 파일
`bot/internal/bot/expert.go`

## 작업 내용

### 메인 핸들러
```go
func (b *Bot) HandleExpert(c telebot.Context) error {
    args := strings.Fields(c.Text())
    if len(args) < 2 {
        return b.expertHelp(c)
    }

    switch args[1] {
    case "list":
        return b.expertList(c)
    case "status":
        return b.expertStatus(c)
    case "ask":
        if len(args) < 3 {
            return c.Send("사용법: /expert ask <expert-name>")
        }
        return b.expertAskStart(c, args[2])
    default:
        return b.expertHelp(c)
    }
}
```

### /expert list
```go
func (b *Bot) expertList(c telebot.Context) error {
    // Expert 목록
    // 활성/비활성 상태
    // 담당 태스크 수
}
```

### /expert status
```go
func (b *Bot) expertStatus(c telebot.Context) error {
    // Expert별 태스크 현황
}
```

### /expert ask (대화형)
```go
func (b *Bot) expertAskStart(c telebot.Context, expertName string) error {
    // Expert 존재 확인
    b.state.SetTempData(c.Sender().ID, "expert", expertName)
    b.state.SetWaiting(c.Sender().ID, WaitingExpertQuestion)
    return c.Send(fmt.Sprintf("%s에게 질문할 내용을 입력하세요:", expertName))
}
```

## 완료 조건
- [ ] HandleExpert 구현
- [ ] expertList 구현
- [ ] expertStatus 구현
- [ ] expertAsk 대화형 구현
