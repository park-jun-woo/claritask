# TASK-BOT-005: 미들웨어

## 목표
인증, 감사 로그, Rate Limiting 미들웨어 구현

## 파일
`bot/internal/bot/middleware.go`

## 작업 내용

### 인증 미들웨어
```go
func (b *Bot) AuthMiddleware() telebot.MiddlewareFunc {
    return func(next telebot.HandlerFunc) telebot.HandlerFunc {
        return func(c telebot.Context) error {
            userID := c.Sender().ID
            if !b.cfg.IsAllowed(userID) {
                b.logger.Warn().
                    Int64("user_id", userID).
                    Str("username", c.Sender().Username).
                    Msg("unauthorized access attempt")
                return c.Send("❌ 권한이 없습니다.")
            }
            return next(c)
        }
    }
}
```

### 관리자 전용 미들웨어
```go
func (b *Bot) AdminOnlyMiddleware() telebot.MiddlewareFunc
```

### 감사 로그 미들웨어
```go
func (b *Bot) AuditMiddleware() telebot.MiddlewareFunc {
    return func(next telebot.HandlerFunc) telebot.HandlerFunc {
        return func(c telebot.Context) error {
            start := time.Now()
            err := next(c)
            b.logger.Info().
                Int64("user_id", c.Sender().ID).
                Str("username", c.Sender().Username).
                Str("text", c.Text()).
                Dur("duration", time.Since(start)).
                Err(err).
                Msg("command")
            return err
        }
    }
}
```

### Rate Limit 미들웨어
```go
func (b *Bot) RateLimitMiddleware() telebot.MiddlewareFunc
```

## 완료 조건
- [ ] AuthMiddleware 구현
- [ ] AdminOnlyMiddleware 구현
- [ ] AuditMiddleware 구현
- [ ] RateLimitMiddleware 구현
