# TASK-BOT-014: 설정 핸들러

## 목표
/settings 명령어 구현

## 파일
`bot/internal/bot/settings.go`

## 작업 내용

### 메인 핸들러
```go
func (b *Bot) HandleSettings(c telebot.Context) error {
    args := strings.Fields(c.Text())
    if len(args) < 2 {
        return b.settingsShow(c)
    }

    switch args[1] {
    case "notify":
        return b.settingsNotify(c)
    case "project":
        return b.settingsProject(c)
    default:
        return b.settingsShow(c)
    }
}
```

### /settings (기본)
```go
func (b *Bot) settingsShow(c telebot.Context) error {
    // 현재 설정 표시
    // 인라인 버튼으로 변경 가능
}
```

### /settings notify
```go
func (b *Bot) settingsNotify(c telebot.Context) error {
    // 알림 설정
    // 인라인 버튼: 태스크 완료, 실패, 새 메시지, 일일 요약
}
```

### 설정 콜백 핸들러
```go
func (b *Bot) handleSettingsCallback(c telebot.Context, parts []string) error {
    // settings:notify:complete:on
    // settings:notify:complete:off
    // settings:notify:fail:on
    // ...
}
```

### 사용자별 설정 저장
```go
type UserSettings struct {
    NotifyOnComplete bool
    NotifyOnFail     bool
    NotifyOnMessage  bool
    DailySummary     bool
}
```

## 완료 조건
- [ ] HandleSettings 구현
- [ ] settingsNotify 인라인 버튼 구현
- [ ] 설정 저장/로드 구현
