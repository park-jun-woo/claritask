# TASK-BOT-008: 라우터

## 목표
명령어 라우팅 설정

## 파일
`bot/internal/bot/router.go`

## 작업 내용

### SetupRoutes 함수
```go
func (b *Bot) SetupRoutes() {
    // 공통 미들웨어
    b.tg.Use(b.AuditMiddleware())
    b.tg.Use(b.RateLimitMiddleware())
    b.tg.Use(b.AuthMiddleware())

    // 기본 명령어
    b.tg.Handle("/start", b.HandleStart)
    b.tg.Handle("/help", b.HandleHelp)
    b.tg.Handle("/status", b.HandleStatus)

    // 프로젝트
    b.tg.Handle("/project", b.HandleProject)

    // 태스크
    b.tg.Handle("/task", b.HandleTask)

    // 메시지
    b.tg.Handle("/msg", b.HandleMessage)

    // Expert
    b.tg.Handle("/expert", b.HandleExpert)

    // 설정
    b.tg.Handle("/settings", b.HandleSettings)

    // 콜백 쿼리 (인라인 버튼)
    b.tg.Handle(telebot.OnCallback, b.HandleCallback)

    // 일반 텍스트 (대화형 입력)
    b.tg.Handle(telebot.OnText, b.HandleText)
}
```

## 완료 조건
- [ ] SetupRoutes 함수 구현
- [ ] 모든 명령어 라우팅 등록
- [ ] 콜백 및 텍스트 핸들러 등록
