# Claribot

English | [한국어](README.ko.md)

LLM-based project automation system. Automate project tasks by controlling Claude Code via Telegram bot and CLI.

## Features

- **Telegram Bot Interface**: Manage projects and run Claude from mobile
- **CLI Client**: Same functionality from terminal
- **Multi-Project Management**: Independent DB per project for task management
- **Claude Code Integration**: PTY-based Claude Code execution and result return
- **Task-Based Workflow**: Message → Task conversion → Execution → Report
- **Cron Scheduler**: Automated Claude execution and result notifications at scheduled times

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    claribot (daemon)                    │
│                                                         │
│  ┌───────────┐  ┌───────────┐  ┌───────────────────┐   │
│  │ Telegram  │  │    CLI    │  │    Scheduler      │   │
│  │  Handler  │  │  Handler  │  │     (cron)        │   │
│  └─────┬─────┘  └─────┬─────┘  └─────────┬─────────┘   │
│        │              │                   │             │
│        └──────────────┼───────────────────┘             │
│                       ▼                                 │
│               ┌──────────────┐                          │
│               │ Router/Claude│                          │
│               └──────────────┘                          │
└─────────────────────────────────────────────────────────┘
         ▲                              ▲
         │ Bot API                      │ HTTP POST
    [Telegram]                     [clari CLI]
```

| Component | Role |
|-----------|------|
| claribot | Telegram + CLI handler, Claude session management (systemd service) |
| clari | HTTP client CLI |
| Claude Code | Performs actual work in project folder |

## Requirements

- Go 1.21+
- Claude Code CLI (`~/.local/bin/claude`)
- SQLite3

## Installation

```bash
# Clone
git clone https://github.com/your-username/claribot.git
cd claribot

# Build and install
make install
```

`make install` performs the following:
- Installs `clari` CLI to `/usr/local/bin/`
- Registers and starts `claribot` service with systemd

### Manual Installation

```bash
# Build only
make build

# Install CLI only
make install-cli

# Install service only
make install-bot
```

## Configuration

Config file: `~/.claribot/config.yaml`

```yaml
# HTTP Service
service:
  host: 127.0.0.1
  port: 9847

# Telegram Bot
telegram:
  token: "YOUR_BOT_TOKEN"    # Get from @BotFather
  allowed_users: []          # Empty = allow all, [123456789] = specific users only
  admin_chat_id: 0           # Schedule execution result notification target (0 = disabled)

# Claude Code
claude:
  timeout: 1200              # Idle timeout (seconds)
  max: 3                     # Maximum concurrent executions

# Project
project:
  path: ~/projects           # Default project creation path

# Pagination
pagination:
  page_size: 10

# Logging
log:
  level: info                # debug, info, warn, error
  file: ~/.claribot/claribot.log
```

Example file: `deploy/config.example.yaml`

## Usage

### Service Management

```bash
make status     # Check service status
make restart    # Restart service
make logs       # View logs (journalctl)
```

### CLI Commands

```bash
# Project management
clari project list              # List projects
clari project create <id>       # Create new project
clari project add <path>        # Register existing folder as project
clari project switch <id>       # Select project
clari project delete <id>       # Delete project

# Task management
clari task list                 # List tasks
clari task add <title>          # Add task
clari task get <id>             # Task details
clari task run [id]             # Run task

# Message (Claude execution)
clari send "Review the code"    # Send message → Run Claude
clari message list              # Message history
clari message status            # Message status summary

# Schedule management
clari schedule list             # List schedules
clari schedule add "0 7 * * *" "Morning greeting"  # Add schedule
clari schedule add --once "30 14 * * *" "One-time notification"  # Run once
clari schedule get <id>         # Schedule details
clari schedule enable <id>      # Enable
clari schedule disable <id>     # Disable
clari schedule delete <id>      # Delete
clari schedule runs <id>        # Execution history
clari schedule set project <id> <project>  # Change project

# Status
clari status                    # Current project status
```

### Telegram Commands

| Command | Description |
|---------|-------------|
| `/start` | Start bot, show menu keyboard |
| `/project` | Project list (selection buttons) |
| `/task` | Task list |
| `/status` | Current status |
| Regular message | Run Claude in selected project |

## Project Structure

```
claribot/
├── bot/                    # claribot service
│   ├── cmd/claribot/       # Entry point
│   ├── internal/           # Internal packages
│   │   ├── config/         # Config loader
│   │   ├── db/             # SQLite wrapper
│   │   ├── handler/        # Command router
│   │   ├── project/        # Project management
│   │   ├── task/           # Task management
│   │   ├── message/        # Message processing
│   │   ├── schedule/       # Schedule management
│   │   ├── prompts/        # System prompt templates
│   │   └── tghandler/      # Telegram handler
│   └── pkg/                # Public packages
│       ├── claude/         # Claude Code execution
│       ├── telegram/       # Telegram Bot API
│       ├── render/         # Markdown → HTML
│       ├── logger/         # Logging
│       └── errors/         # Error types
├── cli/                    # clari CLI
│   └── cmd/clari/
├── deploy/                 # Deployment files
│   ├── claribot.service.template
│   └── config.example.yaml
└── Makefile
```

## Database

### Global DB (`~/.claribot/db.clt`)

Manages projects, schedules, and messages

```sql
projects (
    id TEXT PRIMARY KEY,
    name TEXT,
    path TEXT UNIQUE,
    type TEXT,
    description TEXT,
    status TEXT,
    created_at, updated_at
)

schedules (
    id INTEGER PRIMARY KEY,
    project_id TEXT,          -- NULL for global execution
    cron_expr TEXT,
    message TEXT,
    enabled INTEGER,
    run_once INTEGER,         -- Auto-disable after one-time execution
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
    project_id TEXT,          -- NULL for global execution
    content TEXT,
    source TEXT,              -- 'telegram', 'cli', 'schedule'
    status TEXT,
    result TEXT,
    error TEXT,
    created_at, completed_at
)
```

### Local DB (`<project>/.claribot/db.clt`)

Project-specific tasks (can be managed with git)

```sql
tasks (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER,
    title TEXT,
    spec TEXT,                -- Requirements specification
    plan TEXT,                -- Execution plan
    report TEXT,              -- Completion report
    status TEXT,              -- 'todo', 'planned', 'split', 'done', 'failed'
    error TEXT,
    created_at, updated_at
)
```

## Development

```bash
# Local execution
make run-bot    # Run claribot
make run-cli    # Run CLI

# Test
make test

# Clean build
make clean && make build
```

## Uninstall

```bash
make uninstall
```

## Disclaimer

This project requires Anthropic's [Claude Code](https://claude.ai/claude-code) CLI.

Claribot is a wrapper program that calls Claude Code as a subprocess. It does not include or redistribute Claude Code itself, and users must install Claude Code separately and have an Anthropic account.

**User Responsibility:**
- Users are responsible for complying with [Anthropic Terms of Service](https://www.anthropic.com/legal)
- Automated use on Consumer plans (Free/Pro/Max) may be restricted according to the terms
- For commercial use, reviewing the [Commercial Terms](https://www.anthropic.com/legal/commercial-terms) is recommended

The developers of this project are not responsible for users' violations of Anthropic terms.

## License

MIT License - see [LICENSE](LICENSE)
