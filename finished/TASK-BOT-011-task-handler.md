# TASK-BOT-011: 태스크 핸들러

## 목표
/task 명령어 구현

## 파일
`bot/internal/bot/task.go`

## 작업 내용

### 메인 핸들러
```go
func (b *Bot) HandleTask(c telebot.Context) error {
    args := strings.Fields(c.Text())
    if len(args) < 2 {
        return b.taskHelp(c)
    }

    switch args[1] {
    case "list":
        return b.taskList(c, args[2:])
    case "add":
        return b.taskAddStart(c)
    case "get":
        return b.taskGet(c, args[2:])
    case "start":
        return b.taskStart(c, args[2:])
    case "done":
        return b.taskDone(c, args[2:])
    case "fail":
        return b.taskFail(c, args[2:])
    default:
        return b.taskHelp(c)
    }
}
```

### /task list
```go
func (b *Bot) taskList(c telebot.Context, args []string) error {
    // 상태 필터 (pending, in_progress, completed)
    // 페이지네이션
    // 인라인 버튼 (더 보기)
}
```

### /task add (대화형)
```go
func (b *Bot) taskAddStart(c telebot.Context) error {
    b.state.SetWaiting(c.Sender().ID, WaitingTaskTitle)
    return c.Send("태스크 제목을 입력하세요:")
}

func (b *Bot) handleTaskTitleInput(c telebot.Context, state *UserState) error {
    state.TempData["title"] = c.Text()
    b.state.SetWaiting(c.Sender().ID, WaitingTaskDescription)
    return c.Send("설명을 입력하세요 (스킵: /skip):")
}
```

### /task get
```go
func (b *Bot) taskGet(c telebot.Context, args []string) error {
    // 태스크 상세 정보
    // 인라인 버튼 (완료, 실패, 상세보기)
}
```

### /task start, done, fail
```go
func (b *Bot) taskStart(c telebot.Context, args []string) error
func (b *Bot) taskDone(c telebot.Context, args []string) error
func (b *Bot) taskFail(c telebot.Context, args []string) error
```

### 콜백 핸들러
```go
func (b *Bot) handleTaskCallback(c telebot.Context, parts []string) error {
    // task:start:<id>
    // task:done:<id>
    // task:fail:<id>
    // task:list:page:<n>
}
```

## 완료 조건
- [ ] HandleTask 구현
- [ ] taskList 구현 (페이지네이션)
- [ ] taskAdd 대화형 구현
- [ ] taskStart/Done/Fail 구현
- [ ] 콜백 핸들러 구현
