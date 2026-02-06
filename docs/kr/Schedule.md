# 스케줄 시스템 설계

> Claribot의 Cron 기반 스케줄링 시스템

---

## 개요

지정된 시간에 자동으로 Claude Code 또는 bash 명령을 실행하고 결과를 저장하는 스케줄링 기능. `robfig/cron` 라이브러리를 사용하여 claribot 데몬 내부에서 스케줄을 관리한다.

**핵심 결정**: 외부 cron 대신 내부 루틴 사용
- 통합 관리 (CLI/텔레그램/Web UI로 제어 가능)
- 동적 추가/삭제 가능
- DB 기반 영속성
- 실행 이력 관리

---

## 데이터 구조

### Schedule (스케줄 정의)

```go
type Schedule struct {
    ID        int     `json:"id"`
    ProjectID *string `json:"project_id,omitempty"` // NULL이면 전역
    CronExpr  string  `json:"cron_expr"`            // "0 7 * * *" (매일 오전 7시)
    Message   string  `json:"message"`              // Claude Code 프롬프트 / bash 명령어
    Type      string  `json:"type"`                 // 'claude' (기본) | 'bash'
    Enabled   bool    `json:"enabled"`              // 활성화 여부
    RunOnce   bool    `json:"run_once"`             // 1회 실행 후 자동 비활성화
    LastRun   *string `json:"last_run,omitempty"`   // 마지막 실행 시간
    NextRun   *string `json:"next_run,omitempty"`   // 다음 예정 실행 시간
    CreatedAt string  `json:"created_at"`
    UpdatedAt string  `json:"updated_at"`
}
```

### ScheduleRun (실행 결과)

```go
type ScheduleRun struct {
    ID          int     `json:"id"`
    ScheduleID  int     `json:"schedule_id"`          // 스케줄 ID
    Status      string  `json:"status"`               // 'running', 'done', 'failed'
    Result      string  `json:"result"`               // Claude Code 실행 결과 (리포트)
    Error       string  `json:"error,omitempty"`       // 에러 메시지
    StartedAt   string  `json:"started_at"`           // 실행 시작 시간
    CompletedAt *string `json:"completed_at,omitempty"` // 실행 완료 시간
}
```

### Cron 표현식

```
┌───────────── 분 (0-59)
│ ┌───────────── 시 (0-23)
│ │ ┌───────────── 일 (1-31)
│ │ │ ┌───────────── 월 (1-12)
│ │ │ │ ┌───────────── 요일 (0-6, 일요일=0)
│ │ │ │ │
* * * * *
```

| 예시 | 설명 |
|------|------|
| `0 7 * * *` | 매일 07:00 |
| `30 9 * * 1-5` | 평일 09:30 |
| `0 */2 * * *` | 2시간마다 |
| `0 0 1 * *` | 매월 1일 00:00 |

---

## DB 스키마

### schedules (스케줄 정의)

```sql
CREATE TABLE IF NOT EXISTS schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT,
    cron_expr TEXT NOT NULL,
    message TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'claude'
        CHECK(type IN ('claude', 'bash')),
    enabled INTEGER DEFAULT 1,
    run_once INTEGER DEFAULT 0,
    last_run TEXT,
    next_run TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled);
CREATE INDEX IF NOT EXISTS idx_schedules_project ON schedules(project_id);
```

### schedule_runs (실행 결과)

```sql
CREATE TABLE IF NOT EXISTS schedule_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_id INTEGER NOT NULL,
    status TEXT DEFAULT 'running'
        CHECK(status IN ('running', 'done', 'failed')),
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    started_at TEXT NOT NULL,
    completed_at TEXT,
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_schedule_runs_schedule ON schedule_runs(schedule_id);
CREATE INDEX IF NOT EXISTS idx_schedule_runs_status ON schedule_runs(status);
```

---

## CLI 명령어

### 스케줄 관리
```bash
# 스케줄 추가 (기본 타입: claude)
clari schedule add "0 7 * * *" "오늘의 할 일 정리"
clari schedule add --project claribot "0 9 * * 1-5" "코드 품질 리포트 생성"
clari schedule add --once "30 14 * * *" "5분 후 알림 테스트"  # 1회 실행 후 자동 비활성화

# 스케줄 목록 (타입 컬럼 표시)
clari schedule list              # 현재 프로젝트 스케줄
clari schedule list --all        # 전체 스케줄
clari schedule list --project <id>  # 특정 프로젝트 스케줄

# 스케줄 조회 (타입 표시)
clari schedule get <id>
clari schedule get               # id 없으면: 전체 스케줄 표시

# 스케줄 삭제
clari schedule delete <id>

# 활성화/비활성화
clari schedule enable <id>
clari schedule disable <id>

# 프로젝트 변경
clari schedule set <id> project <project_id>   # 스케줄의 프로젝트 변경
clari schedule set <id> project none            # 전역 실행으로 전환
```

> **참고**: bash 스케줄을 위한 `--type` 옵션은 텔레그램 핸들러에서 지원됩니다 (`schedule add --type bash "*/5 * * * *" "curl -s https://example.com/health"`). CLI는 REST API body로 `type` 필드를 전달하며 기본값은 `claude`입니다.

### 실행 이력
```bash
# 특정 스케줄의 실행 이력
clari schedule runs <schedule_id> [-p <page>] [-n <page_size>]

# 특정 실행 결과 상세 조회
clari schedule run <run_id>
```

---

## REST API

### 엔드포인트

| 메소드 | 엔드포인트 | 설명 |
|--------|----------|------|
| GET | `/api/schedules` | 스케줄 목록 |
| POST | `/api/schedules` | 스케줄 생성 |
| GET | `/api/schedules/{id}` | 스케줄 상세 |
| PATCH | `/api/schedules/{id}` | 스케줄 수정 (field: `project`) |
| DELETE | `/api/schedules/{id}` | 스케줄 삭제 |
| POST | `/api/schedules/{id}/enable` | 스케줄 활성화 |
| POST | `/api/schedules/{id}/disable` | 스케줄 비활성화 |
| GET | `/api/schedules/{id}/runs` | 실행 이력 목록 |
| GET | `/api/schedule-runs/{runId}` | 단건 실행 상세 |

### 쿼리 파라미터

**GET /api/schedules**
- `all=true` - 전체 스케줄 표시 (기본: 현재 프로젝트만)
- `project_id=<id>` - 특정 프로젝트 필터 (`none`이면 전역만)
- `page=<n>`, `page_size=<n>` - 페이지네이션

### 요청/응답 예시

```json
// POST /api/schedules
{
  "cron_expr": "0 9 * * *",
  "message": "일일 리포트",
  "type": "claude",           // 선택, 기본값 "claude"
  "project_id": "blog",       // 선택, NULL이면 전역
  "run_once": false
}

// 응답 (201 Created)
{
  "success": true,
  "message": "스케줄 추가됨: #1\nCron: ...",
  "data": { /* Schedule 객체 */ }
}
```

```json
// PATCH /api/schedules/{id}
{
  "field": "project",
  "value": "blog"         // 또는 "none"으로 전역 전환
}
```

---

## 실행 플로우

### 시작 시
```
[claribot 시작]
    └─ 스케줄러 초기화 (Init)
    └─ 고착 schedule_runs 복구 (running > 1시간 → failed 처리)
    └─ DB에서 enabled=1인 스케줄 로드
    └─ 각 스케줄을 cron에 등록
    └─ cron 시작
    └─ "Scheduler started with N jobs" 로그
```

### 스케줄 실행 시
```
[cron 트리거]
    └─ schedule_runs에 'running' 상태로 레코드 생성
    └─ run_once면 실행 전 자동 비활성화 (재실행 방지)
    └─ project_id로 프로젝트 경로 조회 (없으면 project.DefaultPath)
    └─ 타입별 분기:
    │
    ├─ [type = 'bash']
    │      └─ bash 명령 직접 실행 (5분 타임아웃)
    │      └─ stdout + stderr 결과 캡처
    │      └─ 상태를 'done' 또는 'failed'로 설정
    │
    └─ [type = 'claude'] (기본)
           └─ 리포트 경로 생성 (.claribot/schedule-{runID}-report.md)
           └─ prompts.Get("schedule")에서 시스템 프롬프트 로드
           └─ {{.ReportPath}} 치환으로 템플릿 렌더링
           └─ Claude Code 실행 (메시지를 프롬프트로 전달)
           └─ 인증 오류 확인 (claude.IsAuthError)
           └─ 상태를 'done' 또는 'failed'로 설정
           └─ DB 저장 후 리포트 파일 정리
    │
    └─ schedules의 last_run, next_run 업데이트
    └─ 연속 실패 추적 (성공 시 리셋)
    └─ 3회 연속 실패 → 스케줄 자동 비활성화 + 알림
    └─ 텔레그램으로 결과 알림 전송 (notifier 콜백)
```

### 종료 시
```
[claribot 중지]
    └─ Shutdown()으로 cron 스케줄러 중지
```

### 고착 스케줄 복구
```
[시작 시]
    └─ schedule_runs 조회: status='running' AND started_at < (현재 - 1시간)
    └─ 상태를 'failed', error = 'stuck: recovered on restart'로 업데이트
    └─ 복구된 수 로깅
```

봇이 실행 중 크래시하거나 재시작하면 스케줄이 고착될 수 있다. 복구 로직은 시작 시 자동으로 실행되어 1시간 이상 `running` 상태인 schedule_runs를 `failed`로 표시한다. 타임아웃 상수: `StuckScheduleTimeout = 1 * time.Hour`.

### 연속 실패 자동 비활성화

스케줄이 3회 연속 실패하면 (`MaxConsecutiveFailures = 3`):

1. 스케줄 자동 비활성화 (`enabled = 0`)
2. cron에서 작업 제거
3. 실패 사유와 마지막 오류를 포함한 텔레그램 알림 전송
4. 성공 시 실패 카운터 리셋

### 동적 변경 시
```
[schedule add/delete/enable/disable/set]
    └─ DB 업데이트
    └─ cron에서 작업 추가/제거/재등록
    └─ 활성화 시 next_run 재계산
    └─ 비활성화 시 next_run 초기화
```

---

## 아키텍처

```
┌──────────────────────────────────────────────────────────┐
│                        claribot                           │
│                                                           │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐              │
│  │ 텔레그램 │  │ CLI/REST │  │ 스케줄러  │              │
│  │ 핸들러   │  │ 핸들러   │  │  (cron)   │              │
│  └────┬─────┘  └────┬─────┘  └─────┬─────┘              │
│       │             │              │                     │
│       └──────┬──────┴──────────────┘                     │
│              ▼                                           │
│       ┌──────────────┐                                   │
│       │  타입 체크   │                                   │
│       └──┬───────┬───┘                                   │
│          │       │                                       │
│     claude│      │bash                                   │
│          ▼       ▼                                       │
│   ┌──────────┐ ┌──────────┐                              │
│   │  Claude  │ │   Bash   │  ← 5분 타임아웃             │
│   │   Code   │ │  실행    │                              │
│   └────┬─────┘ └────┬─────┘                              │
│        └──────┬──────┘                                   │
│               ▼                                          │
│   ┌───────────────────────┐                              │
│   │    schedule_runs      │  ← 실행 결과 저장            │
│   │       (DB)            │                              │
│   └───────────┬───────────┘                              │
│               ▼                                          │
│   ┌───────────────────────┐                              │
│   │   Notifier Callback   │  → 텔레그램 알림             │
│   └───────────────────────┘                              │
└──────────────────────────────────────────────────────────┘
```

---

## 구현 파일

```
bot/internal/
├── schedule/
│   ├── schedule.go      # Schedule, ScheduleRun 구조체
│   ├── add.go           # 스케줄 추가 (cron, 프로젝트, 타입 검증)
│   ├── get.go           # 스케줄 상세 조회
│   ├── list.go          # 스케줄 목록 (페이지네이션)
│   ├── delete.go        # 스케줄 삭제 (확인 포함)
│   ├── toggle.go        # 활성화/비활성화 (next_run 재계산)
│   ├── runs.go          # 실행 이력 조회 (페이지네이션)
│   ├── set.go           # 스케줄 속성 변경 (프로젝트)
│   └── scheduler.go     # Cron 매니저 + 실행 로직 + 실패 추적
├── handler/
│   ├── router.go        # 스케줄 명령 (텔레그램/내부, --type 지원)
│   └── restful.go       # 스케줄 REST API 엔드포인트
├── prompts/
│   └── common/
│       └── schedule.md  # claude 타입용 시스템 프롬프트 템플릿 ({{.ReportPath}})
└── db/
    └── db.go            # schedules, schedule_runs 테이블 + 마이그레이션

cli/cmd/clari/
└── main.go              # CLI 스케줄 명령어 (add, list, get, set, delete 등)
```

### 주요 함수

| 함수 | 파일 | 설명 |
|------|------|------|
| `Init(notifier)` | scheduler.go | 전역 스케줄러 초기화, 고착 실행 복구, 작업 로드 |
| `Shutdown()` | scheduler.go | cron 스케줄러 정상 종료 |
| `Register(...)` | scheduler.go | cron에 스케줄 추가/업데이트 (스레드 안전) |
| `Unregister(id)` | scheduler.go | cron에서 스케줄 제거 |
| `execute(...)` | scheduler.go | 예약 작업 실행 (claude 또는 bash) |
| `JobCount()` | scheduler.go | 등록된 cron 작업 수 반환 |

---

## Notifier 콜백

스케줄러는 초기화 시 notifier 콜백 함수를 받아 스케줄 이벤트에 대한 텔레그램 알림을 전송한다.

### 콜백 시그니처
```go
notifier func(projectID *string, msg string)
```

### 초기화
```go
notifier := func(projectID *string, msg string) {
    if bot != nil {
        bot.Broadcast(msg)  // 텔레그램으로 관리자에게 전송
    }
}
schedule.Init(notifier)
```

### 알림 이벤트

| 이벤트 | 이모지 | 형식 |
|--------|--------|------|
| Claude 실행 완료 | `🤖` | `🤖 스케줄 실행 완료: {message}\n\n{result}` |
| Bash 실행 완료 | `🔧` | `🔧 스케줄 실행 완료: {message}\n\n{result}` |
| 실행 실패 | `❌` | `❌{타입이모지} 스케줄 실행 실패: {message}\n\n{error}` |
| 자동 비활성화 (3회 실패) | `⚠️` | `⚠️ 스케줄 자동 비활성화됨\n\n{message}\n\n사유: 3회 연속 실패\n마지막 오류: {error}` |

가독성을 위해 메시지를 잘라서 표시 (message: 50자, result: 500자).

---

## 의존성

```go
import "github.com/robfig/cron/v3"
```

Cron 파서 설정: `cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)` (5필드 형식, 초 없음)

---

## run_once 동작

1회 실행 옵션(`--once`)이 있는 스케줄:

1. cron 시간이 되면 정상적으로 실행
2. Claude Code 실행 **전에** 스케줄 자동 비활성화 (enabled=0)
3. cron에서 작업 제거
4. 실행 결과는 정상적으로 저장

**실행 전 비활성화 이유**: Claude Code 실행 중 오류가 발생해도 재실행 방지

---

## 동시성

- `Scheduler.mu sync.RWMutex`가 `jobs` 맵과 `failureCounts` 맵 보호
- 각 스케줄 실행은 자체 고루틴에서 실행 (cron 라이브러리 관리)
- `Register`와 `Unregister`는 쓰기 잠금 획득
- `JobCount`는 읽기 잠금 획득

---

*Claribot 스케줄 시스템 v0.4*
