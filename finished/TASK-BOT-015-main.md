# TASK-BOT-015: Main 진입점

## 목표
봇 애플리케이션 진입점 구현

## 파일
`bot/cmd/claribot/main.go`

## 작업 내용

### main 함수
```go
package main

import (
    "os"
    "os/signal"
    "syscall"

    "claritask/bot/internal/bot"
    "claritask/bot/internal/config"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func main() {
    // 1. 로거 설정
    setupLogger()

    // 2. 설정 로드
    cfg, err := config.Load()
    if err != nil {
        log.Fatal().Err(err).Msg("failed to load config")
    }

    // 3. 봇 생성
    b, err := bot.New(cfg)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to create bot")
    }

    // 4. 시그널 핸들링
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        log.Info().Msg("shutting down...")
        b.Stop()
    }()

    // 5. 봇 시작
    log.Info().Msg("starting claribot...")
    b.Start()
}

func setupLogger() {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
    log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
```

## 완료 조건
- [ ] main 함수 구현
- [ ] 로거 설정
- [ ] 시그널 핸들링 (graceful shutdown)
