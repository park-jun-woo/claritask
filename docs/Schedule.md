# Schedule System Design

> Claribot's Cron-based Scheduling System

---

## Overview

A scheduling feature that automatically executes Claude Code or bash commands at specified times and stores the results. Manages schedules internally within the claribot daemon using the `robfig/cron` library.

**Key Decision**: Use internal routines instead of external cron
- Unified management (controllable via CLI/Telegram/Web UI)
- Dynamic add/remove capability
- DB-based persistence
- Execution history management

---

## Data Structures

### Schedule (Schedule Definition)

```go
type Schedule struct {
    ID        int     `json:"id"`
    ProjectID *string `json:"project_id,omitempty"` // NULL means global
    CronExpr  string  `json:"cron_expr"`            // "0 7 * * *" (daily at 7 AM)
    Message   string  `json:"message"`              // Prompt for Claude Code / bash command to execute
    Type      string  `json:"type"`                 // 'claude' (default) | 'bash'
    Enabled   bool    `json:"enabled"`              // Whether enabled
    RunOnce   bool    `json:"run_once"`             // Auto-disable after single execution
    LastRun   *string `json:"last_run,omitempty"`   // Last execution time
    NextRun   *string `json:"next_run,omitempty"`   // Next scheduled execution time
    CreatedAt string  `json:"created_at"`
    UpdatedAt string  `json:"updated_at"`
}
```

### ScheduleRun (Execution Result)

```go
type ScheduleRun struct {
    ID          int     `json:"id"`
    ScheduleID  int     `json:"schedule_id"`          // Schedule ID
    Status      string  `json:"status"`               // 'running', 'done', 'failed'
    Result      string  `json:"result"`               // Claude Code execution result (report)
    Error       string  `json:"error,omitempty"`       // Error message
    StartedAt   string  `json:"started_at"`           // Execution start time
    CompletedAt *string `json:"completed_at,omitempty"` // Execution completion time
}
```

### Cron Expression

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-6, Sunday=0)
│ │ │ │ │
* * * * *
```

| Example | Description |
|---------|-------------|
| `0 7 * * *` | Daily at 07:00 |
| `30 9 * * 1-5` | Weekdays at 09:30 |
| `0 */2 * * *` | Every 2 hours |
| `0 0 1 * *` | 1st of every month at 00:00 |

---

## DB Schema

### schedules (Schedule Definitions)

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

### schedule_runs (Execution Results)

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

## CLI Commands

### Schedule Management
```bash
# Add schedule (default type: claude)
clari schedule add "0 7 * * *" "Organize today's todo list"
clari schedule add --project claribot "0 9 * * 1-5" "Generate code quality report"
clari schedule add --once "30 14 * * *" "Notification test in 5 minutes"  # Auto-disable after single execution

# List schedules (shows type column)
clari schedule list              # Current project schedules
clari schedule list --all        # All schedules
clari schedule list --project <id>  # Specific project schedules

# View schedule (shows type)
clari schedule get <id>
clari schedule get               # No id: show all schedules

# Delete schedule
clari schedule delete <id>

# Enable/Disable
clari schedule enable <id>
clari schedule disable <id>

# Change project
clari schedule set <id> project <project_id>   # Change schedule's project
clari schedule set <id> project none            # Switch to global execution
```

> **Note**: The `--type` option for bash schedules is supported via the Telegram handler (`schedule add --type bash "*/5 * * * *" "curl -s https://example.com/health"`). The CLI currently sends the `type` field via the REST API body which defaults to `claude`.

### Execution History
```bash
# Execution history of a specific schedule
clari schedule runs <schedule_id> [-p <page>] [-n <page_size>]

# Detailed view of a specific execution result
clari schedule run <run_id>
```

---

## REST API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/schedules` | List schedules |
| POST | `/api/schedules` | Create new schedule |
| GET | `/api/schedules/{id}` | Get schedule details |
| PATCH | `/api/schedules/{id}` | Update schedule (field: `project`) |
| DELETE | `/api/schedules/{id}` | Delete schedule |
| POST | `/api/schedules/{id}/enable` | Enable schedule |
| POST | `/api/schedules/{id}/disable` | Disable schedule |
| GET | `/api/schedules/{id}/runs` | List execution history |
| GET | `/api/schedule-runs/{runId}` | Get single run details |

### Query Parameters

**GET /api/schedules**
- `all=true` - Show all schedules (default: current project only)
- `project_id=<id>` - Filter by specific project (`none` for global only)
- `page=<n>`, `page_size=<n>` - Pagination

### Request/Response Examples

```json
// POST /api/schedules
{
  "cron_expr": "0 9 * * *",
  "message": "Daily report",
  "type": "claude",           // optional, defaults to "claude"
  "project_id": "blog",       // optional, NULL for global
  "run_once": false
}

// Response (201 Created)
{
  "success": true,
  "message": "스케줄 추가됨: #1\nCron: ...",
  "data": { /* Schedule object */ }
}
```

```json
// PATCH /api/schedules/{id}
{
  "field": "project",
  "value": "blog"         // or "none" to make global
}
```

---

## Execution Flow

### On Startup
```
[claribot starts]
    └─ Initialize Scheduler (Init)
    └─ Recover stuck schedule_runs (running > 1 hour → mark as failed)
    └─ Load schedules with enabled=1 from DB
    └─ Register each schedule with cron
    └─ Start cron
    └─ Log "Scheduler started with N jobs"
```

### On Schedule Execution
```
[cron trigger]
    └─ Create record in schedule_runs with 'running' status
    └─ Auto-disable if run_once (before execution to prevent re-runs)
    └─ Look up project path by project_id (fallback: project.DefaultPath)
    └─ Branch by type:
    │
    ├─ [type = 'bash']
    │      └─ Execute bash command directly (5-minute timeout)
    │      └─ Capture stdout + stderr as result
    │      └─ Set status to 'done' or 'failed'
    │
    └─ [type = 'claude'] (default)
           └─ Generate report path (.claribot/schedule-{runID}-report.md)
           └─ Load system prompt from prompts.Get("schedule")
           └─ Render template with {{.ReportPath}} substitution
           └─ Execute Claude Code (pass message as prompt with system prompt)
           └─ Check for auth errors (claude.IsAuthError)
           └─ Set status to 'done' or 'failed'
           └─ Clean up report file after DB save
    │
    └─ Update last_run, next_run in schedules
    └─ Track consecutive failures (reset on success)
    └─ If 3 consecutive failures → auto-disable schedule + notify
    └─ Send result notification via Telegram (notifier callback)
```

### On Shutdown
```
[claribot stops]
    └─ Shutdown() stops the cron scheduler
```

### Stuck Schedule Recovery
```
[on startup]
    └─ Query schedule_runs WHERE status='running' AND started_at < (now - 1 hour)
    └─ Update status to 'failed', error = 'stuck: recovered on restart'
    └─ Log recovered count
```

Schedules can become stuck if the bot crashes or restarts during execution. The recovery logic runs automatically on startup and marks any schedule_runs that have been in `running` state for more than 1 hour as `failed`. Timeout constant: `StuckScheduleTimeout = 1 * time.Hour`.

### Consecutive Failure Auto-Disable

When a schedule fails 3 consecutive times (`MaxConsecutiveFailures = 3`):

1. The schedule is automatically disabled (`enabled = 0`)
2. The job is unregistered from cron
3. A notification is sent via Telegram with the failure reason and last error message
4. The failure counter resets on any successful execution

### On Dynamic Changes
```
[schedule add/delete/enable/disable/set]
    └─ Update DB
    └─ Add/remove/re-register the job in cron
    └─ Recalculate next_run when enabling
    └─ Clear next_run when disabling
```

---

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                        claribot                           │
│                                                           │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐              │
│  │ Telegram │  │ CLI/REST │  │ Scheduler │              │
│  │ Handler  │  │ Handler  │  │  (cron)   │              │
│  └────┬─────┘  └────┬─────┘  └─────┬─────┘              │
│       │             │              │                     │
│       └──────┬──────┴──────────────┘                     │
│              ▼                                           │
│       ┌──────────────┐                                   │
│       │  Type Check  │                                   │
│       └──┬───────┬───┘                                   │
│          │       │                                       │
│     claude│      │bash                                   │
│          ▼       ▼                                       │
│   ┌──────────┐ ┌──────────┐                              │
│   │  Claude  │ │   Bash   │  ← 5-min timeout            │
│   │   Code   │ │  exec    │                              │
│   └────┬─────┘ └────┬─────┘                              │
│        └──────┬──────┘                                   │
│               ▼                                          │
│   ┌───────────────────────┐                              │
│   │    schedule_runs      │  ← Store execution results   │
│   │       (DB)            │                              │
│   └───────────┬───────────┘                              │
│               ▼                                          │
│   ┌───────────────────────┐                              │
│   │   Notifier Callback   │  → Telegram notification     │
│   └───────────────────────┘                              │
└──────────────────────────────────────────────────────────┘
```

---

## Implementation Files

```
bot/internal/
├── schedule/
│   ├── schedule.go      # Schedule, ScheduleRun structs
│   ├── add.go           # Add schedule (validates cron, project, type)
│   ├── get.go           # View schedule details
│   ├── list.go          # List schedules with pagination
│   ├── delete.go        # Delete schedule (with confirmation)
│   ├── toggle.go        # Enable/Disable (recalculates next_run)
│   ├── runs.go          # View execution history with pagination
│   ├── set.go           # Change schedule properties (project)
│   └── scheduler.go     # Cron manager + execution logic + failure tracking
├── handler/
│   ├── router.go        # Schedule commands (Telegram/internal, supports --type)
│   └── restful.go       # Schedule REST API endpoints
├── prompts/
│   └── common/
│       └── schedule.md  # System prompt template for claude type ({{.ReportPath}})
└── db/
    └── db.go            # schedules, schedule_runs tables + migration

cli/cmd/clari/
└── main.go              # CLI schedule commands (add, list, get, set, delete, etc.)
```

### Key Functions

| Function | File | Description |
|----------|------|-------------|
| `Init(notifier)` | scheduler.go | Initialize global scheduler, recover stuck runs, load jobs |
| `Shutdown()` | scheduler.go | Stop the cron scheduler gracefully |
| `Register(...)` | scheduler.go | Add/update a schedule in cron (thread-safe) |
| `Unregister(id)` | scheduler.go | Remove a schedule from cron |
| `execute(...)` | scheduler.go | Run a scheduled task (claude or bash) |
| `JobCount()` | scheduler.go | Return number of registered cron jobs |

---

## Notifier Callback

The scheduler accepts a notifier callback function on initialization, used to send Telegram notifications for schedule events.

### Callback Signature
```go
notifier func(projectID *string, msg string)
```

### Initialization
```go
notifier := func(projectID *string, msg string) {
    if bot != nil {
        bot.Broadcast(msg)  // Send to admin chat via Telegram
    }
}
schedule.Init(notifier)
```

### Notification Events

| Event | Emoji | Format |
|-------|-------|--------|
| Claude execution complete | `🤖` | `🤖 스케줄 실행 완료: {message}\n\n{result}` |
| Bash execution complete | `🔧` | `🔧 스케줄 실행 완료: {message}\n\n{result}` |
| Execution failed | `❌` | `❌{type_emoji} 스케줄 실행 실패: {message}\n\n{error}` |
| Auto-disabled (3 failures) | `⚠️` | `⚠️ 스케줄 자동 비활성화됨\n\n{message}\n\n사유: 3회 연속 실패\n마지막 오류: {error}` |

Messages are truncated for readability (message: 50 chars, result: 500 chars).

---

## Dependencies

```go
import "github.com/robfig/cron/v3"
```

Cron parser configuration: `cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)` (5-field format, no seconds)

---

## run_once Behavior

Schedules with the one-time execution option (`--once`):

1. Execute normally when the cron time is reached
2. Auto-disable the schedule (enabled=0) **before** Claude Code execution
3. Remove the job from cron
4. Execution results are stored normally

**Reason for disabling before execution**: Prevents re-execution even if an error occurs during Claude Code execution

---

## Concurrency

- `Scheduler.mu sync.RWMutex` protects `jobs` map and `failureCounts` map
- Each schedule execution runs in its own goroutine (managed by cron library)
- `Register` and `Unregister` acquire write lock
- `JobCount` acquires read lock

---

*Claribot Schedule System v0.4*
