# TASK-BOT-010: 프로젝트 핸들러

## 목표
/project 명령어 구현

## 파일
`bot/internal/bot/project.go`

## 작업 내용

### 메인 핸들러
```go
func (b *Bot) HandleProject(c telebot.Context) error {
    args := strings.Fields(c.Text())
    if len(args) < 2 {
        return b.projectHelp(c)
    }

    switch args[1] {
    case "list":
        return b.projectList(c)
    case "status":
        return b.projectStatus(c)
    case "switch":
        if len(args) < 3 {
            return c.Send("사용법: /project switch <project-id>")
        }
        return b.projectSwitch(c, args[2])
    case "info":
        return b.projectInfo(c)
    default:
        return b.projectHelp(c)
    }
}
```

### /project list
```go
func (b *Bot) projectList(c telebot.Context) error {
    projects, err := b.svc.ListProjects()
    // 프로젝트 목록 포맷팅
    // 현재 프로젝트 표시 (⭐)
    // 각 프로젝트 진행률 표시
}
```

### /project status
```go
func (b *Bot) projectStatus(c telebot.Context) error {
    // 현재 프로젝트 상태
    // 진행률 바
    // 태스크 현황 (완료/진행중/대기)
    // 최근 완료/현재 진행 태스크
}
```

### /project switch
```go
func (b *Bot) projectSwitch(c telebot.Context, projectID string) error {
    // 프로젝트 존재 확인
    // 상태 업데이트
    // 확인 메시지
}
```

### 콜백 핸들러
```go
func (b *Bot) handleProjectCallback(c telebot.Context, parts []string) error {
    // project:switch:<id>
    // project:info:<id>
}
```

## 완료 조건
- [ ] HandleProject 구현
- [ ] projectList 구현
- [ ] projectStatus 구현
- [ ] projectSwitch 구현
- [ ] 콜백 핸들러 구현
