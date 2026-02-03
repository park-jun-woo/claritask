# TASK-BOT-001: 프로젝트 초기화

## 목표
bot/ 폴더에 Go 모듈 초기화 및 기본 구조 생성

## 작업 내용

### 1. 디렉토리 구조 생성
```
bot/
├── cmd/claribot/
├── internal/
│   ├── bot/
│   ├── config/
│   └── service/
└── deploy/
```

### 2. go.mod 초기화
```bash
cd bot
go mod init claritask/bot
```

### 3. 의존성 추가
- `gopkg.in/telebot.v3` - 텔레그램 봇 API
- `github.com/caarlos0/env/v9` - 환경변수 파싱
- `github.com/rs/zerolog` - 구조화된 로깅

### 4. go.mod 예상 내용
```go
module claritask/bot

go 1.21

require (
    gopkg.in/telebot.v3 v3.2.1
    github.com/caarlos0/env/v9 v9.0.0
    github.com/rs/zerolog v1.31.0
    modernc.org/sqlite v1.28.0
)
```

## 완료 조건
- [ ] bot/ 디렉토리 구조 생성
- [ ] go.mod 파일 생성
- [ ] go mod tidy 성공
