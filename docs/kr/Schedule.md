# Schedule System Design

> Claribot's Cron-based Scheduling System

---

## Overview

A scheduling feature that automatically executes Claude Code at specified times and stores the results. Manages schedules internally within the claribot daemon using the `robfig/cron` library.

**Key Decision**: Use internal routines instead of external cron
- Unified management (controllable via CLI/Telegram)
- Dynamic add/remove capability
- DB-based persistence
- Execution history management

---

## Data Structures

### Schedule (Schedule Definition)

```go
Schedule {
    ID         int
    ProjectID  *string   // Project ID (NULL means global)
    CronExpr   string    // "0 7 * * *" (daily at 7 AM)
    Message    string    // Prompt to pass to Claude Code
    Enabled    bool      // Whether enabled
    RunOnce    bool      // Auto-disable after single execution
    LastRun    *string   // Last execution time
    NextRun    *string   // Next scheduled execution time
    CreatedAt  string
    UpdatedAt  string
}
```

### ScheduleRun (Execution Result)

```go
ScheduleRun {
    ID          int
    ScheduleID  int       // Schedule ID
    Status      string    // 'running', 'done', 'failed'
    Result      string    // Claude Code execution result (report)
    Error       string    // Error message
    StartedAt   string    // Execution start time
    CompletedAt *string   // Execution completion time
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
# Add schedule
schedule add "0 7 * * *" "Organize today's todo list"
schedule add --project claribot "0 9 * * 1-5" "Generate code quality report"
schedule add --once "30 14 * * *" "Notification test in 5 minutes"  # Auto-disable after single execution

# List schedules
schedule list              # Current project schedules
schedule list --all        # All schedules

# View schedule
schedule get <id>

# Delete schedule
schedule delete <id>

# Enable/Disable
schedule enable <id>
schedule disable <id>

# Change project
schedule set project <id> <project_id>   # Change schedule's project
schedule set project <id> none           # Switch to global execution
```

### Execution History
```bash
# Execution history of a specific schedule
schedule runs <schedule_id>

# Detailed view of a specific execution result
schedule run <run_id>
```

---

## Execution Flow

### On Startup
```
[claribot starts]
    └─ Initialize Scheduler
    └─ Load schedules with enabled=1 from DB
    └─ Register each schedule with cron
    └─ Start cron
```

### On Schedule Execution
```
[cron trigger]
    └─ Create record in schedule_runs with 'running' status
    └─ Retrieve the schedule's message
    └─ Look up project path by project_id
    └─ Execute Claude Code (pass message as prompt)
    └─ Store execution result (report) in schedule_runs
    └─ Update status to 'done' or 'failed'
    └─ Update last_run, next_run in schedules
    └─ Send result notification via Telegram
```

### On Dynamic Changes
```
[schedule add/delete/enable/disable]
    └─ Update DB
    └─ Add/remove the job from cron
```

---

## Architecture

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
│       │  schedule_runs   │  ← Store execution results│
│       │     (DB)         │                          │
│       └──────────────────┘                          │
└─────────────────────────────────────────────────────┘
```

---

## Implementation Files

```
bot/internal/
├── schedule/
│   ├── schedule.go      # Schedule, ScheduleRun structs
│   ├── add.go           # Add schedule
│   ├── get.go           # View schedule
│   ├── list.go          # List schedules
│   ├── delete.go        # Delete schedule
│   ├── toggle.go        # Enable/Disable
│   ├── runs.go          # View execution history
│   ├── set.go           # Change schedule properties
│   └── scheduler.go     # Cron manager + execution logic
├── handler/
│   └── router.go        # Schedule commands
└── db/
    └── db.go            # schedules, schedule_runs tables
```

---

## Dependencies

```go
import "github.com/robfig/cron/v3"
```

---

## run_once Behavior

Schedules with the one-time execution option (`--once`):

1. Execute normally when the cron time is reached
2. Auto-disable the schedule (enabled=0) **before** Claude Code execution
3. Remove the job from cron
4. Execution results are stored normally

**Reason for disabling before execution**: Prevents re-execution even if an error occurs during Claude Code execution

---

*Claribot Schedule System v0.2*
