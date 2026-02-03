# Claribot Overview

> **현재 버전**: v0.0.1 ([변경이력](../HISTORY.md))

---

## 개요

Claribot은 Claritask의 텔레그램 봇 인터페이스입니다. 텔레그램을 통해 프로젝트 상태 확인, 태스크 관리, 메시지 수신/발신 기능을 제공합니다.

**바이너리**: `claribot`
**기술 스택**: Go + telebot.v3 + SQLite

---

## 핵심 기능

| 기능 | 설명 |
|------|------|
| 프로젝트 조회 | 현재 프로젝트 상태 및 진행률 확인 |
| 태스크 관리 | 태스크 목록, 상태 변경, 할당 |
| 메시지 연동 | CLI message와 양방향 연동 |
| 알림 발송 | 태스크 완료/실패 시 자동 알림 |
| Expert 호출 | 특정 Expert에게 질문/지시 전달 |

---

## 아키텍처 개요

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Telegram   │────▶│  Claribot   │────▶│  Claritask  │
│   Server    │◀────│  (Daemon)   │◀────│     DB      │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │  clari CLI  │
                    │  (옵션)     │
                    └─────────────┘
```

### 동작 방식

1. **Long Polling**: 텔레그램 서버에서 메시지 수신
2. **명령어 파싱**: `/task`, `/project` 등 명령어 해석
3. **DB 직접 접근**: 같은 SQLite DB 사용 (읽기 위주)
4. **CLI 호출**: 쓰기 작업은 `clari` CLI 호출 (선택적)

---

## 프로젝트 구조

```
claritask/
├── bot/                         # 텔레그램 봇 소스코드
│   ├── cmd/claribot/
│   │   └── main.go              # 봇 진입점
│   ├── internal/
│   │   ├── bot/                 # 텔레그램 핸들러
│   │   │   ├── bot.go           # Bot 구조체 및 초기화
│   │   │   ├── router.go        # 라우터 설정
│   │   │   ├── middleware.go    # 인증/감사 미들웨어
│   │   │   ├── state.go         # 대화 상태 관리
│   │   │   ├── handlers.go      # 기본 핸들러 (/start, /help, /status)
│   │   │   ├── project.go       # 프로젝트 핸들러
│   │   │   ├── task.go          # 태스크 핸들러
│   │   │   ├── message.go       # 메시지 핸들러
│   │   │   ├── expert.go        # Expert 핸들러
│   │   │   └── settings.go      # 설정 핸들러
│   │   ├── config/              # 설정 관리
│   │   │   └── config.go
│   │   └── service/             # DB 서비스 (CLI 재사용)
│   │       └── service.go
│   ├── deploy/
│   │   ├── claribot.service     # systemd 서비스 파일
│   │   └── claribot.env.example # 환경변수 예시
│   ├── go.mod
│   ├── go.sum
│   └── Makefile
```

---

## 의존성

| 패키지 | 버전 | 용도 |
|--------|------|------|
| gopkg.in/telebot.v3 | v3.2+ | 텔레그램 봇 API |
| modernc.org/sqlite | v1.28+ | SQLite (Pure Go) |

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [02-Architecture.md](02-Architecture.md) | 상세 아키텍처 |
| [03-Commands.md](03-Commands.md) | 봇 명령어 레퍼런스 |
| [04-Deployment.md](04-Deployment.md) | 배포 및 운영 |
| [05-Security.md](05-Security.md) | 보안 설정 |
| [CLI/15-Message.md](../CLI/15-Message.md) | Message 연동 |

---

*Claribot Specification v0.0.1*
