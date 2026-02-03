# Claribot 텔레그램 봇 구현

> 2026-02-04

## 요약

Claritask CLI와 연동되는 텔레그램 봇(Claribot)을 설계하고 구현 완료.

---

## 1. 아키텍처 결정

### 질문
> claritask가 텔레그램 메세지를 받고 소통하려면 systemctl 서비스로 만드는게 맞나?

### 결론
- **systemctl 서비스가 적합** - 24/7 상주, 자동 재시작, journalctl 로그 통합
- 별도 바이너리 `claribot`으로 분리
- 기존 CLI와 동일한 SQLite DB 직접 접근

---

## 2. 명세서 작성 (specs/BOT/)

| 파일 | 내용 |
|------|------|
| 01-Overview.md | 전체 개요, 프로젝트 구조 |
| 02-Architecture.md | 계층별 설계, 데이터 흐름 |
| 03-Commands.md | 봇 명령어 레퍼런스 |
| 04-Deployment.md | systemctl/Docker 배포 |
| 05-Security.md | 인증, Rate Limiting, 감사 로그 |

---

## 3. TASK 계획 (17개)

```
TASK-BOT-001: 프로젝트 초기화 (go.mod)
TASK-BOT-002: Config 모듈
TASK-BOT-003: Service 모듈 (DB 접근)
TASK-BOT-004: Bot 코어
TASK-BOT-005: 미들웨어 (인증/감사/RateLimit)
TASK-BOT-006: 상태 관리
TASK-BOT-007: Rate Limiter
TASK-BOT-008: 라우터
TASK-BOT-009: 기본 핸들러 (/start, /help, /status)
TASK-BOT-010: 프로젝트 핸들러
TASK-BOT-011: 태스크 핸들러
TASK-BOT-012: 메시지 핸들러
TASK-BOT-013: Expert 핸들러
TASK-BOT-014: 설정 핸들러
TASK-BOT-015: Main 진입점
TASK-BOT-016: Makefile
TASK-BOT-017: 배포 파일 (systemd, env)
```

---

## 4. 구현 완료

### 디렉토리 구조
```
bot/
├── cmd/claribot/main.go
├── internal/
│   ├── bot/
│   │   ├── bot.go, router.go, middleware.go
│   │   ├── state.go, ratelimit.go, handlers.go
│   │   ├── project.go, task.go, message.go
│   │   ├── expert.go, settings.go
│   ├── config/config.go
│   └── service/service.go
├── deploy/
│   ├── claribot.service
│   ├── claribot.env.example
│   └── install.sh
├── go.mod, go.sum, Makefile
└── bin/claribot (13.3MB 빌드 성공)
```

### 기술 스택
- Go 1.21+
- gopkg.in/telebot.v3 (텔레그램 API)
- modernc.org/sqlite (Pure Go SQLite)
- github.com/rs/zerolog (로깅)
- github.com/caarlos0/env/v9 (환경변수)

### 지원 명령어
| 명령어 | 설명 |
|--------|------|
| /start | 봇 시작, 현재 프로젝트 표시 |
| /help | 명령어 도움말 |
| /status | 전체 대시보드 |
| /project list/status/switch | 프로젝트 관리 |
| /task list/add/get/start/done/fail | 태스크 관리 |
| /msg list/send | 메시지 관리 |
| /expert list/status/ask | Expert 호출 |
| /settings | 설정 확인 |

### 핵심 기능
1. **인증**: ALLOWED_USERS, ADMIN_USERS 환경변수로 허용 목록 관리
2. **Rate Limiting**: 사용자별 요청 제한 (기본 1/sec, burst 5)
3. **대화형 입력**: 태스크 추가, 메시지 전송 시 대화형 흐름
4. **인라인 버튼**: 프로젝트 전환, 태스크 상태 변경 등
5. **감사 로그**: 모든 명령 실행 zerolog로 기록

---

## 5. 사용법

```bash
# 환경변수 설정
cp bot/deploy/claribot.env.example bot/.env
vim bot/.env  # TELEGRAM_TOKEN, ALLOWED_USERS 등 설정

# 개발 실행
cd bot && make dev

# 프로덕션 배포
sudo ./deploy/install.sh
sudo cp deploy/claribot.env.example /etc/claribot/claribot.env
sudo vim /etc/claribot/claribot.env
make install
sudo systemctl enable claribot
sudo systemctl start claribot
```

---

## 참고

- 텔레그램 봇 토큰: @BotFather에서 발급
- 사용자 ID 확인: @userinfobot 사용
- DB 경로: CLI와 동일한 `~/.claritask/db.clt` 사용
