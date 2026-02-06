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
│              │    DB    │   │ RESTful  │                   │
│              │ (SQLite) │   │   API    │                   │
│              └──────────┘   │ + Auth   │                   │
│                             │ + Legacy │                   │
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
| Auth | JWT + TOTP 2FA authentication | Middleware |
| Claude Code (global) | Runs in ~/.claribot/, router | TTY → kill |
| Claude Code (project) | Runs in project directory, actual work | TTY → kill |

---

## Project Structure

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
│   │   ├── hooks/                # Custom hooks (TanStack Query)
│   │   ├── api/                  # API client (RESTful)
│   │   └── types/                # TypeScript types
│   ├── package.json
│   └── vite.config.ts
├── bot/
│   ├── cmd/claribot/main.go
│   ├── internal/
│   │   ├── auth/                 # JWT + TOTP 2FA authentication
│   │   ├── config/               # YAML config + DB config (key-value)
│   │   ├── db/                   # SQLite wrapper, migrations
│   │   ├── handler/              # RESTful router + handlers
│   │   ├── project/              # Project CRUD + switching
│   │   ├── task/                 # Task CRUD, traversal, cycle
│   │   ├── message/              # Message processing
│   │   ├── schedule/             # Cron-based scheduling
│   │   ├── spec/                 # Requirements specification CRUD
│   │   ├── prompts/              # Prompt templates (file-based)
│   │   ├── tghandler/            # Telegram command handlers
│   │   ├── types/                # Shared types (Result, etc.)
│   │   └── webui/                # Go embed + static file serving
│   └── pkg/
│       ├── claude/               # Claude Code PTY, runner interface, usage stats
│       ├── telegram/             # Telegram bot client
│       ├── render/               # Markdown → HTML rendering
│       ├── logger/               # Structured logging
│       ├── pagination/           # Pagination utilities
│       └── errors/               # Error types
└── cli/
    └── cmd/clari/main.go
```

---

## DB Schema

### Global DB (`~/.claribot/db.clt`)

```sql
projects (
    id TEXT PRIMARY KEY,          -- 'blog', 'api-server'
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,    -- Project path
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
    project_id TEXT,              -- Project ID (NULL = global)
    cron_expr TEXT NOT NULL,      -- "0 7 * * *"
    message TEXT NOT NULL,        -- Prompt or command to execute
    type TEXT NOT NULL DEFAULT 'claude'
        CHECK(type IN ('claude', 'bash')),
    enabled INTEGER DEFAULT 1,
    run_once INTEGER DEFAULT 0,   -- Auto-disable after single execution
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
    project_id TEXT,              -- Project ID (NULL = global)
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
    key TEXT PRIMARY KEY,         -- e.g. 'claude.max', 'project.parallel'
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL
)

auth (
    key TEXT PRIMARY KEY,         -- 'password_hash', 'totp_secret', 'jwt_secret', 'setup_completed'
    value TEXT NOT NULL
)
```

### Local DB (`<project>/.claribot/db.clt`)

```sql
tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER,            -- NULL = top-level Task
    title TEXT NOT NULL,
    spec TEXT DEFAULT '',         -- Requirements specification
    plan TEXT DEFAULT '',         -- Execution plan
    report TEXT DEFAULT '',       -- Execution result report
    status TEXT DEFAULT 'todo'
        CHECK(status IN ('todo', 'split', 'planned', 'done', 'failed')),
    error TEXT DEFAULT '',
    is_leaf INTEGER DEFAULT 1,   -- Leaf node flag
    depth INTEGER DEFAULT 0,     -- Tree depth
    priority INTEGER DEFAULT 0,  -- Execution order (higher = first)
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
)

traversals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK(type IN ('plan', 'run', 'cycle')),
    target_id INTEGER,            -- Root task ID (NULL = all)
    trigger TEXT DEFAULT '',      -- What triggered the traversal
    status TEXT DEFAULT 'running'
        CHECK(status IN ('running', 'done', 'failed', 'cancelled')),
    total INTEGER DEFAULT 0,      -- Total tasks to process
    success INTEGER DEFAULT 0,
    failed INTEGER DEFAULT 0,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    FOREIGN KEY (target_id) REFERENCES tasks(id) ON DELETE SET NULL
)

config (
    key TEXT PRIMARY KEY,         -- e.g. 'parallel'
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

## Configuration

### File Location

```
~/.claribot/
├── config.yaml              -- Service configuration (YAML)
├── db.clt                   -- Global DB (SQLite)
└── claude-usage.txt         -- Cached Claude usage stats
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
  timeout: 1200              # Idle timeout (seconds, default: 1200)
  max_timeout: 1800          # Absolute timeout (seconds, default: 1800)
  max: 10                    # Maximum concurrent executions (default: 10)

project:
  path: "~/projects"         # Default path for project creation

pagination:
  page_size: 10              # Default page size

log:
  level: info                # debug, info, warn, error
  file: ""                   # Log file path (empty = stdout only)
```

### DB Config (Runtime Key-Value)

Runtime configuration stored in the `config` table (both global and local DB). Managed via CLI (`clari config`) or REST API.

Examples:
- `claude.max` — Max concurrent Claude instances (global)
- `parallel` — Project-level parallel execution count (local)

---

## Authentication

### JWT + TOTP 2FA

Claribot uses a password + TOTP (Time-based One-Time Password) two-factor authentication system with JWT tokens.

**Flow**:

```
1. Initial Setup (POST /api/auth/setup)
   ├── Step 1: Send password → Get TOTP URI (for QR code)
   └── Step 2: Send password + TOTP code → Get JWT token

2. Login (POST /api/auth/login)
   └── Send password + TOTP code → Get JWT token

3. Authenticated Requests
   └── JWT stored in HTTP-only cookie (claribot_token)
```

**Components** (`bot/internal/auth/`):
- `auth.go` — Setup, Login, Status, token validation
- `jwt.go` — JWT token generation/validation (24h expiry, auto-generated secret)
- `totp.go` — TOTP secret generation and validation (SHA1, 6 digits, 30s period)
- `password.go` — bcrypt password hashing and verification

**Storage**: Global DB `auth` table (key-value: `password_hash`, `totp_secret`, `jwt_secret`, `setup_completed`)

---

## Spec System

Requirements specification management per project. Specs are stored in the local DB `specs` table.

**Lifecycle**: `draft` → `review` → `approved` → `deprecated`

**Operations**: CRUD via CLI (`clari spec`), Telegram, REST API, and Web UI (Specs page).

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

- `/start` → Exposes `!project` button on Reply keyboard + permanent keyboard with common commands
- `/status` → Shows per-project task progress and active cycle status
- `/usage` → Shows Claude Code usage statistics
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

### TTY Management

- **Idle timeout**: Auto-kill after configured timeout (default: 1200s)
- **Max timeout**: Absolute timeout regardless of activity (default: 1800s)
- **Health check**: Background goroutine monitors and cleans up stale sessions
- **Parallel execution**: Configurable per-project concurrent Claude instances
- **Auth error detection**: Automatically stops traversal on Claude authentication failures

---

## Claude Usage Statistics

Claribot provides Claude Code usage monitoring:

- **CLI**: `clari usage` — Shows stats from `stats-cache.json`
- **Telegram**: `/usage` — Shows usage with live rate limit info
- **REST API**: `GET /api/usage` — Returns stats + cached live usage
- **Live refresh**: `POST /api/usage/refresh` — Triggers async rate limit check

---

## RESTful API Endpoints

All API endpoints follow RESTful conventions. Authentication required via JWT cookie (except `/api/auth/*`). Localhost requests bypass auth.

### Auth

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/auth/status` | Check setup/auth status |
| POST | `/api/auth/setup` | Initial setup (password + TOTP) |
| POST | `/api/auth/login` | Login (password + TOTP → JWT) |
| POST | `/api/auth/logout` | Logout (clear JWT cookie) |
| GET | `/api/auth/totp-setup` | Get TOTP provisioning URI |

### Status & Usage

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/status` | Service status, cycle status, task stats |
| GET | `/api/usage` | Claude Code usage statistics |
| POST | `/api/usage/refresh` | Trigger live usage refresh |
| GET | `/api/health` | Service health (version, uptime, claude slots) |
| GET/POST | `/api` | Legacy API (backward compat for CLI/Telegram, `?args=` query) |

### Projects

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/projects` | List projects (paginated) |
| POST | `/api/projects` | Create/add project |
| GET | `/api/projects/stats` | Per-project task statistics |
| GET | `/api/projects/{id}` | Get project details |
| PATCH | `/api/projects/{id}` | Update project fields |
| DELETE | `/api/projects/{id}` | Delete project |
| POST | `/api/projects/{id}/switch` | Switch active project |

### Tasks

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/tasks` | List tasks (paginated, ?tree=true for full tree) |
| POST | `/api/tasks` | Create task |
| GET | `/api/tasks/{id}` | Get task details (spec, plan, report) |
| PATCH | `/api/tasks/{id}` | Update task field |
| DELETE | `/api/tasks/{id}` | Delete task |
| POST | `/api/tasks/{id}/plan` | Plan single task |
| POST | `/api/tasks/{id}/run` | Run single task |
| POST | `/api/tasks/plan-all` | Plan all pending tasks |
| POST | `/api/tasks/run-all` | Run all planned tasks |
| POST | `/api/tasks/cycle` | Full cycle (plan + run iteratively) |
| POST | `/api/tasks/stop` | Stop active traversal |

### Messages

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/messages` | List messages (paginated) |
| POST | `/api/messages` | Send message |
| GET | `/api/messages/{id}` | Get message details |
| GET | `/api/messages/status` | Message queue status |
| GET | `/api/messages/processing` | Currently processing messages |

### Configs (DB Key-Value)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/configs` | List all config entries |
| GET | `/api/configs/{key}` | Get config value |
| PUT | `/api/configs/{key}` | Set config value |
| DELETE | `/api/configs/{key}` | Delete config entry |

### Config YAML (Raw File)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/config-yaml` | Read config.yaml content |
| PUT | `/api/config-yaml` | Write config.yaml content |

### Schedules

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/schedules` | List schedules (paginated) |
| POST | `/api/schedules` | Create schedule (type: claude/bash) |
| GET | `/api/schedules/{id}` | Get schedule details |
| PATCH | `/api/schedules/{id}` | Update schedule field |
| DELETE | `/api/schedules/{id}` | Delete schedule |
| POST | `/api/schedules/{id}/enable` | Enable schedule |
| POST | `/api/schedules/{id}/disable` | Disable schedule |
| GET | `/api/schedules/{id}/runs` | List schedule runs |
| GET | `/api/schedule-runs/{runId}` | Get run details |

### Specs

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/specs` | List specs (paginated) |
| POST | `/api/specs` | Create spec |
| GET | `/api/specs/{id}` | Get spec details |
| PATCH | `/api/specs/{id}` | Update spec field |
| DELETE | `/api/specs/{id}` | Delete spec |

### Static Files

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/*` | Web UI static file serving (SPA fallback) |

---

## Implementation Status

### Core Infrastructure
- [x] Makefile, systemd service configuration
- [x] DB schema (global/local) with auto-migration
- [x] Configuration validation
- [x] Logging system (structured, file + stdout)
- [x] Graceful shutdown
- [x] Error handling (rows.Err, LastInsertId, os.MkdirAll)

### Authentication
- [x] JWT + TOTP 2FA authentication (`bot/internal/auth/`)
- [x] Auth middleware for API endpoints
- [x] Setup and Login API endpoints
- [x] Frontend Login/Setup pages with routing guard

### Telegram
- [x] Telegram package (buttons, callbacks)
- [x] Chat ID authentication (allowed_users)
- [x] Project switching (inline buttons)
- [x] Permanent keyboard with common commands
- [x] Message status header (current project display)
- [x] Report: summary inline + detailed file attachment
- [x] HTML formatted messages
- [x] `/status` with per-project progress and cycle status
- [x] `/usage` Claude Code usage statistics
- [x] Goroutine limit and pendingContext cleanup

### Claude Code Integration
- [x] Claude Code PTY integration
- [x] Task execution and result return
- [x] Idle timeout + absolute timeout (max_timeout)
- [x] Stale session health check goroutine
- [x] Auth error detection and traversal stop
- [x] Prompt templates (file-based, extracted from hardcoded)
- [x] Usage statistics (stats-cache.json + live rate limit)

### Task System
- [x] Task CRUD with tree structure (parent_id, depth, is_leaf)
- [x] Task status: todo → split/planned → done/failed
- [x] Task priority-based execution ordering
- [x] Plan/Run/Cycle traversal with parallel execution
- [x] Traversal tracking table (traversals)
- [x] Traversal stop command
- [x] Context Map based context injection
- [x] Report file auto-cleanup

### Project Management
- [x] Project CRUD (create, add, list, get, set, delete, switch)
- [x] Project rollback on create failure
- [x] Per-project parallel execution setting
- [x] Per-project task statistics API (`/api/projects/stats`)

### Schedule System
- [x] Cron-based scheduling
- [x] Schedule types: `claude` (Claude Code) and `bash` (direct command)
- [x] Schedule execution result notification (Telegram)
- [x] Schedule run history

### Message System
- [x] Message processing (telegram, cli, gui, schedule sources)
- [x] Global execution support
- [x] Context Map based context injection

### Config System
- [x] YAML config file (config.yaml) with validation
- [x] DB config table (key-value, runtime settings)
- [x] Config CRUD API + CLI commands
- [x] Raw YAML read/write API

### Spec System
- [x] Spec CRUD (add, list, get, set, delete)
- [x] Spec lifecycle: draft → review → approved → deprecated
- [x] REST API endpoints
- [x] CLI commands (`clari spec`)
- [x] Telegram handlers
- [x] Web UI Specs page

### Web UI
- [x] React + shadcn/ui + TanStack Query (Go embed)
- [x] RESTful API client
- [x] Dashboard with per-project task stats + cycle progress
- [x] Projects page with create/edit/switch
- [x] Tasks page with tree view, status bar, plan/run/cycle controls
- [x] Messages page (chat UI with markdown rendering)
- [x] Schedules page (type: claude/bash)
- [x] Specs page
- [x] Settings page (config YAML editor)
- [x] Login/Setup pages (2FA authentication)
- [x] Mobile responsive (hamburger menu, touch targets, card views)
- [x] SPA fallback routing
- [x] CORS middleware (development)

### REST API
- [x] RESTful router (`bot/internal/handler/restful.go`)
- [x] Full CRUD endpoints for all resources
- [x] Pagination support
- [x] Auth middleware integration

---

*Claribot v0.2.21*
