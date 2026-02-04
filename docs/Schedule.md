# Schedule 시스템 설계

> Claribot의 Cron 기반 스케줄링 시스템

---

## 개요

지정된 시간에 자동으로 Claude Code를 실행하고 결과를 저장하는 스케줄링 기능. `robfig/cron` 라이브러리를 활용하여 claribot 데몬 내부에서 스케줄을 관리한다.

**핵심 결정**: 외부 cron 대신 자체 루틴 사용
- 통합 관리 (CLI/텔레그램으로 제어)
- 동적 추가/삭제 가능
- DB 기반 영속성
- 실행 결과 이력 관리

---

## 데이터 구조

### Schedule (스케줄 정의)

```go
Schedule {
    ID         int
    ProjectID  string    // 프로젝트 ID (NULL이면 전역)
    CronExpr   string    // "0 7 * * *" (매일 7시)
    Message    string    // Claude Code에 전달할 프롬프트
    Enabled    bool      // 활성화 여부
    LastRun    *string   // 마지막 실행 시간
    NextRun    *string   // 다음 실행 예정 시간
    CreatedAt  string
    UpdatedAt  string
}
```

### ScheduleRun (실행 결과)

```go
ScheduleRun {
    ID          int
    ScheduleID  int       // 스케줄 ID
    Status      string    // 'running', 'done', 'failed'
    Result      string    // Claude Code 실행 결과 (보고서)
    Error       string    // 에러 메시지
    StartedAt   string    // 실행 시작 시간
    CompletedAt *string   // 실행 완료 시간
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
    enabled INTEGER DEFAULT 1,
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
# 스케줄 추가
schedule add "0 7 * * *" "오늘의 할일 목록을 정리해줘"
schedule add --project claribot "0 9 * * 1-5" "코드 품질 리포트 작성해줘"

# 스케줄 목록
schedule list              # 현재 프로젝트 스케줄
schedule list --all        # 전체 스케줄

# 스케줄 조회
schedule get <id>

# 스케줄 삭제
schedule delete <id>

# 활성화/비활성화
schedule enable <id>
schedule disable <id>
```

### 실행 기록 조회
```bash
# 특정 스케줄의 실행 기록
schedule runs <schedule_id>

# 특정 실행 결과 상세 조회
schedule run <run_id>
```

---

## 실행 흐름

### 시작 시
```
[claribot 시작]
    └─ Scheduler 초기화
    └─ DB에서 enabled=1인 스케줄 로드
    └─ 각 스케줄을 cron에 등록
    └─ cron 시작
```

### 스케줄 실행 시
```
[cron 트리거]
    └─ schedule_runs에 'running' 상태로 레코드 생성
    └─ 해당 스케줄의 message 조회
    └─ project_id로 프로젝트 경로 조회
    └─ Claude Code 실행 (message를 프롬프트로 전달)
    └─ 실행 결과(보고서)를 schedule_runs에 저장
    └─ status를 'done' 또는 'failed'로 업데이트
    └─ schedules의 last_run, next_run 업데이트
    └─ 텔레그램으로 결과 알림 전송
```

### 동적 변경 시
```
[schedule add/delete/enable/disable]
    └─ DB 업데이트
    └─ cron에서 해당 job 추가/제거
```

---

## 아키텍처

```
┌─────────────────────────────────────────────────────┐
│                     claribot                         │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐         │
│  │ Telegram │  │   CLI    │  │ Scheduler │         │
│  │ Handler  │  │ Handler  │  │  (cron)   │         │
│  └────┬─────┘  └────┬─────┘  └─────┬─────┘         │
│       │             │              │                │
│       └──────┬──────┴──────────────┘                │
│              ▼                                      │
│       ┌──────────┐                                  │
│       │  Claude  │                                  │
│       │   Code   │                                  │
│       └────┬─────┘                                  │
│            │                                        │
│       ┌────▼─────────────┐                          │
│       │  schedule_runs   │  ← 실행 결과 저장        │
│       │     (DB)         │                          │
│       └──────────────────┘                          │
└─────────────────────────────────────────────────────┘
```

---

## 구현 파일

```
bot/internal/
├── schedule/
│   ├── schedule.go      # Schedule, ScheduleRun 구조체
│   ├── add.go           # 스케줄 추가
│   ├── get.go           # 스케줄 조회
│   ├── list.go          # 스케줄 목록
│   ├── delete.go        # 스케줄 삭제
│   ├── toggle.go        # 활성화/비활성화
│   ├── runs.go          # 실행 기록 조회
│   └── scheduler.go     # cron 매니저 + 실행 로직
├── handler/
│   └── router.go        # schedule 명령어
└── db/
    └── db.go            # schedules, schedule_runs 테이블
```

---

## 의존성

```go
import "github.com/robfig/cron/v3"
```

---

*Claribot Schedule System v0.2*
