# Claribot v0.2

> **버전**: v0.2.21

---

## 핵심 개념

**Claribot = 장기 주행 컨텍스트 오케스트레이터**

- Claude Code 단독: 컨텍스트 창 안에서 해결 가능한 작업
- Claribot: 컨텍스트 창을 넘어서는 거대 작업의 분할 정복

### Task 기반 분할 정복

1. 사용자가 메시지 입력 (→ 최상위 Task로 등록)
2. claribot이 Claude Code에게 Task 분할 지시
3. Claude Code가 하위 Task들 등록
4. 생성된 Task들을 순서대로 실행
5. Task 수행 중 추가 하위 Task 등록 가능
6. 모든 Task 완료 → 결과 반환

---

## 아키텍처

```
┌──────────────────────────────────────────────────────────┐
│                     claribot (daemon)                     │
│                                                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ Telegram │  │   CLI    │  │  Web UI  │  │   TTY    │ │
│  │ Handler  │  │ Handler  │  │ Handler  │  │ Manager  │ │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘ │
│       │             │             │              │        │
│       └─────────────┼─────────────┼──────────────┘        │
│                     ▼             │                        │
│              ┌──────────┐   ┌────▼─────┐                  │
│              │    DB    │   │ /api     │ ← POST JSON      │
│              └──────────┘   │ /api/health│ ← GET           │
│                             │ /*       │ ← SPA (embed)    │
│                             └──────────┘                  │
└──────────────────────────────────────────────────────────┘
       ▲              ▲              ▲
       │ 메시지        │ HTTP         │ HTTP (브라우저)
  [Telegram]     [clari CLI]     [Web UI]
```

### 컴포넌트

| 컴포넌트 | 역할 | 실행 |
|----------|------|------|
| claribot | Telegram + CLI + WebUI 핸들러 + TTY 관리 | systemctl 서비스 |
| clari CLI | 서비스에 HTTP 요청 | 일시 실행 |
| Web UI | 브라우저 기반 대시보드 (Go embed) | 127.0.0.1:9847 |
| Claude Code (전역) | ~/.claribot/에서 실행, 라우터 | TTY → kill |
| Claude Code (프로젝트) | 프로젝트 폴더에서 실행, 실제 작업 | TTY → kill |

---

## 프로젝트 구조

```
claribot/
├── Makefile
├── deploy/
│   ├── claribot.service.template
│   └── config.example.yaml
├── gui/                          # Web UI (React + TypeScript + Vite)
│   ├── src/
│   │   ├── components/           # Layout, UI 컴포넌트
│   │   ├── pages/                # Dashboard, Projects, Tasks, Messages, Schedules, Settings
│   │   ├── hooks/                # 커스텀 훅 (TanStack Query)
│   │   ├── api/                  # API 클라이언트
│   │   └── types/                # TypeScript 타입
│   ├── package.json
│   └── vite.config.ts
├── bot/
│   ├── cmd/claribot/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── db/
│   │   ├── handler/
│   │   ├── project/
│   │   ├── task/
│   │   ├── message/
│   │   ├── schedule/
│   │   ├── prompts/
│   │   ├── tghandler/
│   │   └── webui/                # Go embed + 정적 파일 서빙
│   └── pkg/
│       ├── claude/
│       ├── telegram/
│       ├── render/
│       ├── logger/
│       └── errors/
└── cli/
    └── cmd/clari/main.go
```

---

## DB 스키마

### 전역 DB (`~/.claribot/db.clt`)

```sql
projects (
    id TEXT PRIMARY KEY,      -- 'blog', 'api-server'
    name TEXT,
    path TEXT UNIQUE,         -- 프로젝트 경로
    type TEXT,                -- 'dev.platform', 'dev.cli', 'write.webnovel'
    description TEXT,
    status TEXT,              -- 'active', 'archived'
    created_at, updated_at
)

schedules (
    id INTEGER PRIMARY KEY,
    project_id TEXT,          -- 프로젝트 ID (NULL이면 전역)
    cron_expr TEXT,           -- "0 7 * * *"
    message TEXT,             -- Claude Code에 전달할 프롬프트
    enabled INTEGER,          -- 활성화 여부
    run_once INTEGER,         -- 1회 실행 후 자동 비활성화
    last_run TEXT,
    next_run TEXT,
    created_at, updated_at
)

schedule_runs (
    id INTEGER PRIMARY KEY,
    schedule_id INTEGER,
    status TEXT,              -- 'running', 'done', 'failed'
    result TEXT,
    error TEXT,
    started_at, completed_at
)

messages (
    id INTEGER PRIMARY KEY,
    project_id TEXT,          -- 프로젝트 ID (NULL이면 전역)
    content TEXT,
    source TEXT,              -- 'telegram', 'cli', 'schedule'
    status TEXT,              -- 'pending', 'processing', 'done', 'failed'
    result TEXT,
    error TEXT,
    created_at, completed_at
)
```

### 로컬 DB (`프로젝트/.claribot/db.clt`)

```sql
tasks (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER,        -- NULL이면 최상위 Task (=사용자 메시지)
    source TEXT,              -- 'telegram', 'cli', 'agent', ''
    title TEXT,
    content TEXT,
    status TEXT,              -- 'pending', 'running', 'done', 'failed'
    result TEXT,
    error TEXT,
    created_at, started_at, completed_at
)
```

**핵심 단순화**: Message를 별도 테이블로 두지 않고 Task로 통합. `parent_id=NULL`이면 최상위 Task.

---

## 설정

### 파일 위치

```
~/.claribot/
├── config.yaml              -- 서비스 설정
└── db.clt                   -- 전역 DB
```

### config.yaml

```yaml
service:
  port: 9847
  host: 127.0.0.1

telegram:
  token: "BOT_TOKEN"
  allowed_users:
    - 123456789
  admin_chat_id: 123456789     # 스케줄 실행 결과 알림 대상

claude:
  timeout: 1200              # idle timeout (초)
  max: 3                     # 동시 실행 최대 개수

project:
  path: "~/projects"         # 프로젝트 생성 기본 경로
```

---

## 설치

```bash
# 빌드 및 설치
make install

# 서비스 관리
make status    # 상태 확인
make restart   # 재시작
make logs      # 로그 확인

# 제거
make uninstall
```

---

## 텔레그램 사용법

### 프로젝트 스위칭

```
1. !project 클릭 → 프로젝트 목록 (Inline 버튼)
2. [claribot] 버튼 클릭 → 프로젝트 선택됨
3. 메시지 입력 → 선택된 프로젝트로 처리
4. 다른 프로젝트로 바꾸려면 !project 다시 클릭
```

- `/start` → Reply 키보드에 `!project` 버튼 노출
- `project.path` → 프로젝트 생성 시 기본 경로 (예: `~/projects/blog`)

---

## Claude Code 실행 모델

### 2-Depth 제한

```
claribot
    └──TTY──▶ Claude Code [전역] (~/.claribot/)
                    └──Task──▶ Claude Code [프로젝트] (프로젝트 경로)
                                    └──▶ 작업 수행 (더 이상 네스트 안 함)
```

---

## 구현 현황

- [x] Makefile, systemd 서비스 설정
- [x] DB 스키마 (전역/로컬)
- [x] Telegram 패키지 (버튼, 콜백)
- [x] claribot 메인 (Telegram 연동)
- [x] 프로젝트 스위칭
- [x] 메시지 → Task 생성 로직
- [x] Claude Code PTY 연동
- [x] Task 실행 및 결과 반환
- [x] clari CLI HTTP 클라이언트 (POST 방식)
- [x] Markdown → HTML 렌더링
- [x] 로깅 시스템
- [x] Graceful shutdown
- [x] 설정 검증
- [x] Schedule 시스템 (cron 기반)
- [x] 스케줄 실행 결과 텔레그램 알림
- [x] Message 전역 실행 지원
- [x] Web UI (React + shadcn/ui, Go embed)
- [x] API 엔드포인트 분리 (`/api`, `/api/health`)
- [x] CORS 미들웨어 (개발용)
- [x] SPA fallback 라우팅

---

## HTTP 엔드포인트

| 엔드포인트 | 메서드 | 용도 |
|-----------|--------|------|
| `POST /` | POST | CLI 하위 호환 (기존) |
| `POST /api` | POST | 커맨드 실행 (Web UI) |
| `GET /api/health` | GET | 서비스 헬스체크 (version, uptime, claude) |
| `GET /*` | GET | Web UI 정적 파일 서빙 (SPA fallback) |

---

*Claribot v0.2.21*
