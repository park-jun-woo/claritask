# Claribot v0.2

> **버전**: v0.2.21

---

## 핵심 개념

**Claribot = 장기 주행 컨텍스트 오케스트레이터**

- Claude Code 단독: 컨텍스트 창 안에서 해결 가능한 작업
- Claribot: 컨텍스트 창을 넘어서는 거대 작업의 분할 정복

### Task 기반 분할 정복

1. 사용자가 메시지 입력 (→ 최상위 Task로 등록)
2. Claribot이 Claude Code에게 Task 분할 지시
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
│              │    DB    │   │ RESTful  │                   │
│              │ (SQLite) │   │   API    │                   │
│              └──────────┘   │ + Auth   │                   │
│                             │ + Legacy │                   │
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
| Auth | JWT + TOTP 2FA 인증 | 미들웨어 |
| Claude Code (전역) | ~/.claribot/에서 실행, 라우터 | TTY → kill |
| Claude Code (프로젝트) | 프로젝트 디렉토리에서 실행, 실제 작업 | TTY → kill |

---

## 프로젝트 구조

```
claribot/
├── Makefile
├── deploy/
│   ├── claribot.service.template
│   ├── claribot-deploy.sh
│   ├── setup-sudoers.sh
│   └── config.example.yaml
├── gui/                          # Web UI (React + TypeScript + Vite)
│   ├── src/
│   │   ├── components/           # Layout, UI, ProjectSelector,
│   │   │                         # ChatBubble, MarkdownRenderer
│   │   ├── pages/                # Dashboard, Projects, ProjectEdit, Tasks,
│   │   │                         # Messages, Schedules, Specs, Settings,
│   │   │                         # Login, Setup
│   │   ├── hooks/                # 커스텀 훅 (TanStack Query)
│   │   ├── api/                  # API 클라이언트 (RESTful)
│   │   └── types/                # TypeScript 타입
│   ├── package.json
│   └── vite.config.ts
├── bot/
│   ├── cmd/claribot/main.go
│   ├── internal/
│   │   ├── auth/                 # JWT + TOTP 2FA 인증
│   │   ├── config/               # YAML 설정 + DB 설정 (key-value)
│   │   ├── db/                   # SQLite 래퍼, 마이그레이션
│   │   ├── handler/              # RESTful 라우터 + 핸들러
│   │   ├── project/              # 프로젝트 CRUD + 전환
│   │   ├── task/                 # Task CRUD, 순회, 사이클
│   │   ├── message/              # 메시지 처리
│   │   ├── schedule/             # Cron 기반 스케줄링
│   │   ├── spec/                 # 요구사항 명세서 CRUD
│   │   ├── prompts/              # 프롬프트 템플릿 (파일 기반)
│   │   ├── tghandler/            # 텔레그램 명령 핸들러
│   │   ├── types/                # 공유 타입 (Result 등)
│   │   └── webui/                # Go embed + 정적 파일 서빙
│   └── pkg/
│       ├── claude/               # Claude Code PTY, 러너 인터페이스, 사용량 통계
│       ├── telegram/             # 텔레그램 봇 클라이언트
│       ├── render/               # Markdown → HTML 렌더링
│       ├── logger/               # 구조화 로깅
│       ├── pagination/           # 페이지네이션 유틸리티
│       └── errors/               # 에러 타입
└── cli/
    └── cmd/clari/main.go
```

---

## DB 스키마

### 전역 DB (`~/.claribot/db.clt`)

```sql
projects (
    id TEXT PRIMARY KEY,          -- 'blog', 'api-server'
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,    -- 프로젝트 경로
    description TEXT DEFAULT '',
    status TEXT DEFAULT 'active'  -- 'active', 'archived'
        CHECK(status IN ('active', 'archived')),
    category TEXT DEFAULT '',
    pinned INTEGER DEFAULT 0,
    last_accessed TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
)

schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT,              -- 프로젝트 ID (NULL = 전역)
    cron_expr TEXT NOT NULL,      -- "0 7 * * *"
    message TEXT NOT NULL,        -- 실행할 프롬프트 또는 명령어
    type TEXT NOT NULL DEFAULT 'claude'
        CHECK(type IN ('claude', 'bash')),
    enabled INTEGER DEFAULT 1,
    run_once INTEGER DEFAULT 0,   -- 1회 실행 후 자동 비활성화
    last_run TEXT,
    next_run TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
)

schedule_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_id INTEGER NOT NULL,
    status TEXT DEFAULT 'running'
        CHECK(status IN ('running', 'done', 'failed')),
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    started_at TEXT NOT NULL,
    completed_at TEXT,
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE
)

messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT,              -- 프로젝트 ID (NULL = 전역)
    content TEXT NOT NULL,
    source TEXT DEFAULT ''
        CHECK(source IN ('', 'telegram', 'cli', 'gui', 'schedule')),
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'processing', 'done', 'failed')),
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    completed_at TEXT
)

config (
    key TEXT PRIMARY KEY,         -- 예: 'claude.max', 'project.parallel'
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL
)

auth (
    key TEXT PRIMARY KEY,         -- 'password_hash', 'totp_secret', 'jwt_secret', 'setup_completed'
    value TEXT NOT NULL
)
```

### 로컬 DB (`<프로젝트>/.claribot/db.clt`)

```sql
tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER,            -- NULL = 최상위 Task
    title TEXT NOT NULL,
    spec TEXT DEFAULT '',         -- 요구사항 명세
    plan TEXT DEFAULT '',         -- 실행 계획
    report TEXT DEFAULT '',       -- 실행 결과 보고서
    status TEXT DEFAULT 'todo'
        CHECK(status IN ('todo', 'split', 'planned', 'done', 'failed')),
    error TEXT DEFAULT '',
    is_leaf INTEGER DEFAULT 1,   -- 리프 노드 플래그
    depth INTEGER DEFAULT 0,     -- 트리 깊이
    priority INTEGER DEFAULT 0,  -- 실행 순서 (낮을수록 먼저)
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
)

traversals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK(type IN ('plan', 'run', 'cycle')),
    target_id INTEGER,            -- 루트 Task ID (NULL = 전체)
    trigger TEXT DEFAULT '',      -- 순회를 트리거한 원인
    status TEXT DEFAULT 'running'
        CHECK(status IN ('running', 'done', 'failed', 'cancelled')),
    total INTEGER DEFAULT 0,      -- 처리할 총 Task 수
    success INTEGER DEFAULT 0,
    failed INTEGER DEFAULT 0,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    FOREIGN KEY (target_id) REFERENCES tasks(id) ON DELETE SET NULL
)

config (
    key TEXT PRIMARY KEY,         -- 예: 'parallel'
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL
)

specs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT DEFAULT '',
    status TEXT DEFAULT 'draft'
        CHECK(status IN ('draft', 'review', 'approved', 'deprecated')),
    priority INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
)
```

---

## 설정

### 파일 위치

```
~/.claribot/
├── config.yaml              -- 서비스 설정 (YAML)
├── db.clt                   -- 전역 DB (SQLite)
└── claude-usage.txt         -- 캐시된 Claude 사용량 통계
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
  timeout: 1200              # 유휴 타임아웃 (초, 기본값: 1200)
  max_timeout: 1800          # 절대 타임아웃 (초, 기본값: 1800)
  max: 10                    # 최대 동시 실행 수 (기본값: 10)

project:
  path: "~/projects"         # 프로젝트 생성 기본 경로

pagination:
  page_size: 10              # 기본 페이지 크기

log:
  level: info                # debug, info, warn, error
  file: ""                   # 로그 파일 경로 (빈 값 = stdout만)
```

### DB Config (런타임 Key-Value)

`config` 테이블(전역/로컬 DB)에 저장되는 런타임 설정. CLI(`clari config`) 또는 REST API로 관리.

예시:
- `claude.max` — 최대 동시 Claude 인스턴스 수 (전역)
- `parallel` — 프로젝트별 병렬 실행 수 (로컬)

---

## 인증

### JWT + TOTP 2FA

Claribot은 비밀번호 + TOTP(시간 기반 일회용 비밀번호) 이중 인증 시스템과 JWT 토큰을 사용합니다.

**플로우**:

```
1. 초기 설정 (POST /api/auth/setup)
   ├── 1단계: 비밀번호 전송 → TOTP URI 반환 (QR 코드용)
   └── 2단계: 비밀번호 + TOTP 코드 전송 → JWT 토큰 발급

2. 로그인 (POST /api/auth/login)
   └── 비밀번호 + TOTP 코드 전송 → JWT 토큰 발급

3. 인증된 요청
   └── JWT는 HTTP-only 쿠키(claribot_token)에 저장
```

**컴포넌트** (`bot/internal/auth/`):
- `auth.go` — Setup, Login, Status, 토큰 검증
- `jwt.go` — JWT 토큰 생성/검증 (24시간 만료, 자동 생성 시크릿)
- `totp.go` — TOTP 시크릿 생성 및 검증 (SHA1, 6자리, 30초 주기)
- `password.go` — bcrypt 비밀번호 해싱 및 검증

**저장소**: 전역 DB `auth` 테이블 (key-value: `password_hash`, `totp_secret`, `jwt_secret`, `setup_completed`)

---

## Spec 시스템

프로젝트별 요구사항 명세서 관리. Spec은 로컬 DB `specs` 테이블에 저장됩니다.

**생명주기**: `draft` → `review` → `approved` → `deprecated`

**조작**: CLI(`clari spec`), 텔레그램, REST API, Web UI(Specs 페이지)를 통한 CRUD.

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

### 프로젝트 전환

```
1. !project 클릭 → 프로젝트 목록 (인라인 버튼)
2. [claribot] 버튼 클릭 → 프로젝트 선택됨
3. 메시지 입력 → 선택된 프로젝트로 처리
4. 다른 프로젝트로 바꾸려면 !project 다시 클릭
```

- `/start` → Reply 키보드에 `!project` 버튼 노출 + 자주 쓰는 명령어 영구 키보드
- `/status` → 프로젝트별 Task 진행상황 및 활성 순회 상태 표시
- `/usage` → Claude Code 사용량 통계 표시
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

### TTY 관리

- **유휴 타임아웃**: 설정된 시간 후 자동 종료 (기본값: 1200초)
- **절대 타임아웃**: 활동 여부와 무관한 절대 타임아웃 (기본값: 1800초)
- **헬스 체크**: 백그라운드 고루틴이 방치된 세션 감시 및 정리
- **병렬 실행**: 프로젝트별 동시 Claude 인스턴스 수 설정 가능
- **인증 오류 감지**: Claude 인증 실패 시 자동으로 순회 중단

---

## Claude 사용량 통계

Claribot은 Claude Code 사용량 모니터링을 제공합니다:

- **CLI**: `clari usage` — `stats-cache.json` 기반 통계 표시
- **텔레그램**: `/usage` — 실시간 속도 제한 정보와 함께 사용량 표시
- **REST API**: `GET /api/usage` — 통계 + 캐시된 실시간 사용량 반환
- **실시간 갱신**: `POST /api/usage/refresh` — 비동기 속도 제한 확인 트리거

---

## RESTful API 엔드포인트

모든 API 엔드포인트는 RESTful 규약을 따릅니다. JWT 쿠키를 통한 인증 필요 (`/api/auth/*` 제외). localhost 요청은 인증 우회.

### Auth

| 메서드 | 엔드포인트 | 용도 |
|--------|----------|------|
| GET | `/api/auth/status` | 설정/인증 상태 확인 |
| POST | `/api/auth/setup` | 초기 설정 (비밀번호 + TOTP) |
| POST | `/api/auth/login` | 로그인 (비밀번호 + TOTP → JWT) |
| POST | `/api/auth/logout` | 로그아웃 (JWT 쿠키 삭제) |
| GET | `/api/auth/totp-setup` | TOTP 프로비저닝 URI 조회 |

### 상태 & 사용량

| 메서드 | 엔드포인트 | 용도 |
|--------|----------|------|
| GET | `/api/status` | 서비스 상태, 순회 상태, Task 통계 |
| GET | `/api/usage` | Claude Code 사용량 통계 |
| POST | `/api/usage/refresh` | 실시간 사용량 갱신 트리거 |
| GET | `/api/health` | 서비스 헬스 (버전, 업타임, Claude 슬롯) |
| GET/POST | `/api` | 레거시 API (CLI/텔레그램 하위 호환, `?args=` 쿼리) |

### Projects

| 메서드 | 엔드포인트 | 용도 |
|--------|----------|------|
| GET | `/api/projects` | 프로젝트 목록 (페이지네이션) |
| POST | `/api/projects` | 프로젝트 생성/추가 |
| GET | `/api/projects/stats` | 프로젝트별 Task 통계 |
| GET | `/api/projects/{id}` | 프로젝트 상세 조회 |
| PATCH | `/api/projects/{id}` | 프로젝트 필드 수정 |
| DELETE | `/api/projects/{id}` | 프로젝트 삭제 |
| POST | `/api/projects/{id}/switch` | 활성 프로젝트 전환 |

### Tasks

| 메서드 | 엔드포인트 | 용도 |
|--------|----------|------|
| GET | `/api/tasks` | Task 목록 (페이지네이션, ?tree=true로 전체 트리) |
| POST | `/api/tasks` | Task 생성 |
| GET | `/api/tasks/{id}` | Task 상세 조회 (spec, plan, report) |
| PATCH | `/api/tasks/{id}` | Task 필드 수정 |
| DELETE | `/api/tasks/{id}` | Task 삭제 |
| POST | `/api/tasks/{id}/plan` | 단일 Task 계획 |
| POST | `/api/tasks/{id}/run` | 단일 Task 실행 |
| POST | `/api/tasks/plan-all` | 대기 중인 모든 Task 계획 |
| POST | `/api/tasks/run-all` | 계획된 모든 Task 실행 |
| POST | `/api/tasks/cycle` | 전체 순회 (plan + run 반복) |
| POST | `/api/tasks/stop` | 활성 순회 중단 |

### Messages

| 메서드 | 엔드포인트 | 용도 |
|--------|----------|------|
| GET | `/api/messages` | 메시지 목록 (페이지네이션) |
| POST | `/api/messages` | 메시지 전송 |
| GET | `/api/messages/{id}` | 메시지 상세 조회 |
| GET | `/api/messages/status` | 메시지 큐 상태 |
| GET | `/api/messages/processing` | 현재 처리 중인 메시지 |

### Configs (DB Key-Value)

| 메서드 | 엔드포인트 | 용도 |
|--------|----------|------|
| GET | `/api/configs` | 전체 설정 목록 |
| GET | `/api/configs/{key}` | 설정값 조회 |
| PUT | `/api/configs/{key}` | 설정값 저장 |
| DELETE | `/api/configs/{key}` | 설정 삭제 |

### Config YAML (원본 파일)

| 메서드 | 엔드포인트 | 용도 |
|--------|----------|------|
| GET | `/api/config-yaml` | config.yaml 내용 조회 |
| PUT | `/api/config-yaml` | config.yaml 내용 저장 |

### Schedules

| 메서드 | 엔드포인트 | 용도 |
|--------|----------|------|
| GET | `/api/schedules` | 스케줄 목록 (페이지네이션) |
| POST | `/api/schedules` | 스케줄 생성 (type: claude/bash) |
| GET | `/api/schedules/{id}` | 스케줄 상세 조회 |
| PATCH | `/api/schedules/{id}` | 스케줄 필드 수정 |
| DELETE | `/api/schedules/{id}` | 스케줄 삭제 |
| POST | `/api/schedules/{id}/enable` | 스케줄 활성화 |
| POST | `/api/schedules/{id}/disable` | 스케줄 비활성화 |
| GET | `/api/schedules/{id}/runs` | 스케줄 실행 이력 |
| GET | `/api/schedule-runs/{runId}` | 실행 상세 조회 |

### Specs

| 메서드 | 엔드포인트 | 용도 |
|--------|----------|------|
| GET | `/api/specs` | Spec 목록 (페이지네이션) |
| POST | `/api/specs` | Spec 생성 |
| GET | `/api/specs/{id}` | Spec 상세 조회 |
| PATCH | `/api/specs/{id}` | Spec 필드 수정 |
| DELETE | `/api/specs/{id}` | Spec 삭제 |

### 정적 파일

| 메서드 | 엔드포인트 | 용도 |
|--------|----------|------|
| GET | `/*` | Web UI 정적 파일 서빙 (SPA fallback) |

---

## 구현 현황

### 코어 인프라
- [x] Makefile, systemd 서비스 설정
- [x] DB 스키마 (전역/로컬) 자동 마이그레이션
- [x] 설정 검증
- [x] 로깅 시스템 (구조화, 파일 + stdout)
- [x] Graceful shutdown
- [x] 에러 처리 (rows.Err, LastInsertId, os.MkdirAll)

### 인증
- [x] JWT + TOTP 2FA 인증 (`bot/internal/auth/`)
- [x] API 엔드포인트용 인증 미들웨어
- [x] Setup 및 Login API 엔드포인트
- [x] 프론트엔드 Login/Setup 페이지 및 라우팅 가드

### 텔레그램
- [x] 텔레그램 패키지 (버튼, 콜백)
- [x] Chat ID 인증 (allowed_users)
- [x] 프로젝트 전환 (인라인 버튼)
- [x] 자주 쓰는 명령어 영구 키보드
- [x] 메시지 상태 헤더 (현재 프로젝트 표시)
- [x] Report: 요약 인라인 + 상세 파일 첨부
- [x] HTML 포맷 메시지
- [x] `/status` 프로젝트별 진행상황 및 순회 상태
- [x] `/usage` Claude Code 사용량 통계
- [x] 고루틴 제한 및 pendingContext 정리

### Claude Code 연동
- [x] Claude Code PTY 연동
- [x] Task 실행 및 결과 반환
- [x] 유휴 타임아웃 + 절대 타임아웃 (max_timeout)
- [x] 방치 세션 헬스 체크 고루틴
- [x] 인증 오류 감지 및 순회 중단
- [x] 프롬프트 템플릿 (파일 기반, 하드코딩 제거)
- [x] 사용량 통계 (stats-cache.json + 실시간 속도 제한)

### Task 시스템
- [x] Task CRUD (트리 구조: parent_id, depth, is_leaf)
- [x] Task 상태: todo → split/planned → done/failed
- [x] Task priority 기반 실행 순서
- [x] Plan/Run/Cycle 순회 (병렬 실행)
- [x] 순회 추적 테이블 (traversals)
- [x] 순회 중단 명령
- [x] Context Map 기반 맥락 주입
- [x] Report 파일 자동 정리

### 프로젝트 관리
- [x] 프로젝트 CRUD (create, add, list, get, set, delete, switch)
- [x] 프로젝트 생성 실패 시 롤백
- [x] 프로젝트별 병렬 실행 설정
- [x] 프로젝트별 Task 통계 API (`/api/projects/stats`)

### 스케줄 시스템
- [x] Cron 기반 스케줄링
- [x] 스케줄 타입: `claude` (Claude Code) 및 `bash` (직접 명령)
- [x] 스케줄 실행 결과 알림 (텔레그램)
- [x] 스케줄 실행 이력

### 메시지 시스템
- [x] 메시지 처리 (telegram, cli, gui, schedule 소스)
- [x] 전역 실행 지원
- [x] Context Map 기반 맥락 주입

### Config 시스템
- [x] YAML 설정 파일 (config.yaml) 검증
- [x] DB config 테이블 (key-value, 런타임 설정)
- [x] Config CRUD API + CLI 명령
- [x] 원본 YAML 읽기/쓰기 API

### Spec 시스템
- [x] Spec CRUD (add, list, get, set, delete)
- [x] Spec 생명주기: draft → review → approved → deprecated
- [x] REST API 엔드포인트
- [x] CLI 명령 (`clari spec`)
- [x] 텔레그램 핸들러
- [x] Web UI Specs 페이지

### Web UI
- [x] React + shadcn/ui + TanStack Query (Go embed)
- [x] RESTful API 클라이언트
- [x] Dashboard (프로젝트별 Task 통계 + 순회 진행률)
- [x] Projects 페이지 (생성/수정/전환)
- [x] Tasks 페이지 (트리뷰, 상태바, plan/run/cycle 제어)
- [x] Messages 페이지 (채팅 UI, 마크다운 렌더링)
- [x] Schedules 페이지 (type: claude/bash)
- [x] Specs 페이지
- [x] Settings 페이지 (config YAML 에디터)
- [x] Login/Setup 페이지 (2FA 인증)
- [x] 모바일 반응형 (햄버거 메뉴, 터치 타겟, 카드뷰)
- [x] SPA fallback 라우팅
- [x] CORS 미들웨어 (개발용)

### REST API
- [x] RESTful 라우터 (`bot/internal/handler/restful.go`)
- [x] 모든 리소스 CRUD 엔드포인트
- [x] 페이지네이션 지원
- [x] 인증 미들웨어 연동

---

*Claribot v0.2.21*
