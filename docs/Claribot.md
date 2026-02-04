# Claribot v0.2

> **버전**: v0.2.19

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

### 핵심 기술: 그래프 기반 컨텍스트 선별 주입

- 전체 히스토리 덤프가 아닌 Edge로 연결된 Task만 주입
- O(n) → O(k) 컨텍스트 사용량 최적화

---

## 아키텍처

```
┌─────────────────────────────────────────────────────────┐
│                    claribot (daemon)                    │
│                                                         │
│  ┌───────────┐  ┌───────────┐  ┌───────────────────┐   │
│  │ Telegram  │  │    CLI    │  │    TTY Manager    │   │
│  │  Handler  │  │  Handler  │  │  (Claude Code)    │   │
│  └─────┬─────┘  └─────┬─────┘  └─────────┬─────────┘   │
│        │              │                   │             │
│        └──────────────┼───────────────────┘             │
│                       ▼                                 │
│               ┌──────────────┐                          │
│               │      DB      │                          │
│               └──────────────┘                          │
└─────────────────────────────────────────────────────────┘
         ▲                              ▲
         │ 메시지                        │ HTTP
    [Telegram]                     [clari CLI]
```

### 컴포넌트

| 컴포넌트 | 역할 | 실행 |
|----------|------|------|
| claribot | Telegram + CLI 핸들러 + TTY 관리 | systemctl 서비스 |
| clari CLI | 서비스에 HTTP 요청 | 일시 실행 |
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
├── bot/
│   ├── cmd/claribot/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── db/
│   │   ├── handler/
│   │   ├── project/
│   │   ├── task/
│   │   ├── message/
│   │   ├── edge/
│   │   └── tghandler/
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
    description TEXT,
    status TEXT,              -- 'active', 'archived'
    created_at, updated_at
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

task_edges (
    from_task_id INTEGER,     -- 선행 작업
    to_task_id INTEGER,       -- 후행 작업
    created_at
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

---

*Claribot v0.2.19*
