# Claribot v0.2

> **Version**: v0.2.21

---

## Core Concept

**Claribot = Long-Running Context Orchestrator**

- Claude Code alone: Tasks solvable within a single context window
- Claribot: Divide-and-conquer for large tasks that exceed a single context window

### Task-Based Divide and Conquer

1. User inputs a message (→ registered as a top-level Task)
2. Claribot instructs Claude Code to split the Task
3. Claude Code registers sub-Tasks
4. Generated Tasks are executed in order
5. Additional sub-Tasks can be registered during Task execution
6. All Tasks complete → results returned

---

## Architecture

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
       │ Message       │ HTTP         │ HTTP (Browser)
  [Telegram]     [clari CLI]     [Web UI]
```

### Components

| Component | Role | Execution |
|-----------|------|-----------|
| claribot | Telegram + CLI + WebUI handler + TTY management | systemctl service |
| clari CLI | HTTP requests to the service | One-shot execution |
| Web UI | Browser-based dashboard (Go embed) | 127.0.0.1:9847 |
| Claude Code (global) | Runs in ~/.claribot/, router | TTY → kill |
| Claude Code (project) | Runs in project directory, actual work | TTY → kill |

---

## Project Structure

```
claribot/
├── Makefile
├── deploy/
│   ├── claribot.service.template
│   └── config.example.yaml
├── gui/                          # Web UI (React + TypeScript + Vite)
│   ├── src/
│   │   ├── components/           # Layout, UI components
│   │   ├── pages/                # Dashboard, Projects, Tasks, Messages, Schedules, Settings
│   │   ├── hooks/                # Custom hooks (TanStack Query)
│   │   ├── api/                  # API client
│   │   └── types/                # TypeScript types
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
│   │   └── webui/                # Go embed + static file serving
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

## DB Schema

### Global DB (`~/.claribot/db.clt`)

```sql
projects (
    id TEXT PRIMARY KEY,      -- 'blog', 'api-server'
    name TEXT,
    path TEXT UNIQUE,         -- Project path
    description TEXT,
    status TEXT,              -- 'active', 'archived'
    created_at, updated_at
)

schedules (
    id INTEGER PRIMARY KEY,
    project_id TEXT,          -- Project ID (NULL = global)
    cron_expr TEXT,           -- "0 7 * * *"
    message TEXT,             -- Prompt to pass to Claude Code
    enabled INTEGER,          -- Enabled flag
    run_once INTEGER,         -- Auto-disable after single execution
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
    project_id TEXT,          -- Project ID (NULL = global)
    content TEXT,
    source TEXT,              -- 'telegram', 'cli', 'schedule'
    status TEXT,              -- 'pending', 'processing', 'done', 'failed'
    result TEXT,
    error TEXT,
    created_at, completed_at
)
```

### Local DB (`<project>/.claribot/db.clt`)

```sql
tasks (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER,        -- NULL = top-level Task (= user message)
    source TEXT,              -- 'telegram', 'cli', 'agent', ''
    title TEXT,
    content TEXT,
    status TEXT,              -- 'pending', 'running', 'done', 'failed'
    result TEXT,
    error TEXT,
    created_at, started_at, completed_at
)
```

**Key simplification**: Messages are not stored in a separate table but unified into Tasks. `parent_id=NULL` indicates a top-level Task.

---

## Configuration

### File Location

```
~/.claribot/
├── config.yaml              -- Service configuration
└── db.clt                   -- Global DB
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
  admin_chat_id: 123456789     # Notification target for schedule execution results

claude:
  timeout: 1200              # Idle timeout (seconds)
  max: 3                     # Maximum concurrent executions

project:
  path: "~/projects"         # Default path for project creation
```

---

## Installation

```bash
# Build and install
make install

# Service management
make status    # Check status
make restart   # Restart
make logs      # View logs

# Uninstall
make uninstall
```

---

## Telegram Usage

### Project Switching

```
1. Click !project → Project list (Inline buttons)
2. Click [claribot] button → Project selected
3. Input message → Processed under the selected project
4. To switch to another project, click !project again
```

- `/start` → Exposes `!project` button on Reply keyboard
- `project.path` → Default path for project creation (e.g., `~/projects/blog`)

---

## Claude Code Execution Model

### 2-Depth Limit

```
claribot
    └──TTY──▶ Claude Code [global] (~/.claribot/)
                    └──Task──▶ Claude Code [project] (project path)
                                    └──▶ Execute work (no further nesting)
```

---

## Implementation Status

- [x] Makefile, systemd service configuration
- [x] DB schema (global/local)
- [x] Telegram package (buttons, callbacks)
- [x] claribot main (Telegram integration)
- [x] Project switching
- [x] Message → Task creation logic
- [x] Claude Code PTY integration
- [x] Task execution and result return
- [x] clari CLI HTTP client (POST method)
- [x] Markdown → HTML rendering
- [x] Logging system
- [x] Graceful shutdown
- [x] Configuration validation
- [x] Schedule system (cron-based)
- [x] Schedule execution result Telegram notification
- [x] Message global execution support
- [x] Web UI (React + shadcn/ui, Go embed)
- [x] API endpoint separation (`/api`, `/api/health`)
- [x] CORS middleware (development)
- [x] SPA fallback routing

---

## HTTP Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `POST /` | POST | CLI backward compatibility (legacy) |
| `POST /api` | POST | Command execution (Web UI) |
| `GET /api/health` | GET | Service health check (version, uptime, claude) |
| `GET /*` | GET | Web UI static file serving (SPA fallback) |

---

*Claribot v0.2.21*
