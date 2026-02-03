# Claritask Code Report

## 1. Overview

| Category | Value |
|----------|-------|
| Language | Go 1.21+ |
| CLI Framework | Cobra v1.8.0 |
| Database | SQLite (mattn/go-sqlite3 v1.14.22) |
| Binary Name | `clari` |
| Total Source Lines | 3,056 lines |
| Total Test Lines | 2,812 lines |
| Test Coverage | 7 test files |

## 2. Project Structure

```
claritask/
├── cmd/claritask/
│   └── main.go              # Entry point (13 lines)
├── internal/
│   ├── cmd/                  # CLI commands (1,629 lines)
│   │   ├── root.go          # Root command, DB/JSON helpers
│   │   ├── init.go          # Project initialization
│   │   ├── project.go       # Project CRUD
│   │   ├── phase.go         # Phase management
│   │   ├── task.go          # Task lifecycle
│   │   ├── memo.go          # Memo operations
│   │   ├── context.go       # Project context
│   │   ├── tech.go          # Tech stack
│   │   ├── design.go        # Design decisions
│   │   └── required.go      # Required check
│   ├── service/             # Business logic (1,174 lines)
│   │   ├── project_service.go
│   │   ├── phase_service.go
│   │   ├── task_service.go
│   │   ├── memo_service.go
│   │   └── state_service.go
│   ├── db/                  # Database layer (140 lines)
│   │   └── db.go
│   └── model/               # Data models (113 lines)
│       └── models.go
└── test/                    # Tests (2,812 lines)
    ├── db_test.go
    ├── models_test.go
    ├── project_service_test.go
    ├── phase_service_test.go
    ├── task_service_test.go
    ├── memo_service_test.go
    └── state_service_test.go
```

## 3. CLI Commands

### Command Tree

```
clari
├── init <project-id> [description]  # Initialize project
├── project
│   ├── set '<json>'                  # Set project config
│   ├── get                           # Get project info
│   ├── plan                          # Start planning mode
│   └── start                         # Start execution mode
├── phase
│   ├── create '<json>'               # Create phase
│   ├── list                          # List phases
│   ├── plan <phase_id>               # Plan phase
│   └── start <phase_id>              # Start phase
├── task
│   ├── push '<json>'                 # Create task
│   ├── pop                           # Get next task (with manifest)
│   ├── start <task_id>               # Start task
│   ├── complete <task_id> '<json>'   # Complete task
│   ├── fail <task_id> '<json>'       # Fail task
│   ├── status                        # Show progress
│   ├── get <task_id>                 # Get task detail
│   └── list <phase_id>               # List tasks by phase
├── memo
│   ├── set <key> '<json>'            # Set memo
│   ├── get <key>                     # Get memo
│   ├── list [scope]                  # List memos
│   └── del <key>                     # Delete memo
├── context
│   ├── set '<json>'                  # Set context
│   └── get                           # Get context
├── tech
│   ├── set '<json>'                  # Set tech stack
│   └── get                           # Get tech stack
├── design
│   ├── set '<json>'                  # Set design
│   └── get                           # Get design
└── required                          # Check required fields
```

## 4. Data Models

### 4.1 Core Models (`internal/model/models.go`)

| Model | Fields | Purpose |
|-------|--------|---------|
| `Project` | ID, Name, Description, Status, CreatedAt | Project metadata |
| `Phase` | ID, ProjectID, Name, Description, OrderNum, Status, CreatedAt | Work phase |
| `Task` | ID, PhaseID, ParentID, Status, Title, Level, Skill, References, Content, Result, Error, timestamps | Atomic work unit |
| `Context` | ID(=1), Data(JSON), timestamps | Project context (singleton) |
| `Tech` | ID(=1), Data(JSON), timestamps | Tech stack (singleton) |
| `Design` | ID(=1), Data(JSON), timestamps | Design decisions (singleton) |
| `State` | Key, Value | Runtime state (key-value) |
| `Memo` | Scope, ScopeID, Key, Data(JSON), Priority, timestamps | Scoped notes |

### 4.2 State Flow

```
Task: pending → doing → done/failed
Phase: pending → active → done
Project: active → archived
```

## 5. Database Schema

### Tables

| Table | Primary Key | Description |
|-------|-------------|-------------|
| `projects` | id (TEXT) | Project registry |
| `phases` | id (INTEGER AUTO) | Phase with FK to project |
| `tasks` | id (INTEGER AUTO) | Task with FK to phase, self-ref parent_id |
| `context` | id (CHECK=1) | Singleton context |
| `tech` | id (CHECK=1) | Singleton tech |
| `design` | id (CHECK=1) | Singleton design |
| `state` | key (TEXT) | Key-value store |
| `memos` | (scope, scope_id, key) | Composite PK |

### Foreign Keys
- `phases.project_id` → `projects.id`
- `tasks.phase_id` → `phases.id`
- `tasks.parent_id` → `tasks.id` (self-reference)

## 6. Code Quality Analysis

### 6.1 Strengths

| Aspect | Details |
|--------|---------|
| **Consistent Structure** | All commands follow same pattern: parse input → open DB → execute → output JSON |
| **Error Handling** | Early return pattern, error wrapping with `fmt.Errorf("context: %w", err)` |
| **JSON Output** | All CLI outputs standardized with `success` field |
| **Separation of Concerns** | Clear layers: cmd (CLI) → service (business logic) → db (persistence) |
| **Time Handling** | Centralized in `db.TimeNow()` and `db.ParseTime()` |
| **Naming Convention** | Consistent Go style (camelCase vars, PascalCase types) |

### 6.2 Areas for Improvement

| Issue | Location | Suggestion |
|-------|----------|------------|
| **Duplicate scan logic** | `task_service.go:71-113`, `task_service.go:128-170` | Extract to shared function for `rows.Scan` |
| **Inconsistent error return** | `cmd/*.go` commands return `nil` after `outputError()` | Consider returning error for proper exit codes |
| **Missing validation** | `phase_service.go:CreatePhase` | No validation for duplicate names |
| **Hardcoded options** | `project_service.go:199-252` | Move required field options to config |
| **No pagination** | `ListTasks`, `ListPhases` | Add LIMIT/OFFSET for large datasets |
| **Unused ParseMemoKey** | `memo_service.go:215-231` | Function defined but not used |

### 6.3 Code Metrics by File

| File | Lines | Complexity |
|------|-------|------------|
| `task.go` | 387 | High (8 subcommands) |
| `task_service.go` | 359 | High (PopTask with manifest) |
| `project_service.go` | 335 | Medium |
| `memo.go` | 254 | Medium |
| `phase.go` | 243 | Medium |
| `memo_service.go` | 231 | Medium |
| `project.go` | 224 | Medium |
| `db.go` | 140 | Low |
| `phase_service.go` | 128 | Low |
| `state_service.go` | 121 | Low |
| `models.go` | 113 | Low (data only) |

## 7. Test Coverage

| Test File | Lines | Coverage Area |
|-----------|-------|---------------|
| `task_service_test.go` | 560 | Task CRUD, status transitions |
| `project_service_test.go` | 501 | Project, Context, Tech, Design |
| `memo_service_test.go` | 491 | Memo CRUD, scope handling |
| `models_test.go` | 366 | Model instantiation |
| `phase_service_test.go` | 316 | Phase CRUD, status |
| `db_test.go` | 305 | DB open, migrate, close |
| `state_service_test.go` | 273 | State CRUD |

### Missing Test Coverage
- CLI commands (`internal/cmd/*`)
- Integration tests (end-to-end workflow)
- Error edge cases

## 8. Key Workflows

### 8.1 Project Initialization (`clari init`)

```
1. Validate project ID (regex: ^[a-z0-9_-]+$)
2. Create directory: ./<project-id>/.claritask/
3. Initialize SQLite DB with schema
4. Create project record
5. Initialize state keys
6. Generate CLAUDE.md template
```

### 8.2 Task Pop (`clari task pop`)

```
1. Query next pending task (ORDER BY id LIMIT 1)
2. Start task (pending → doing)
3. Build manifest:
   - context, tech, design from singletons
   - state from key-value store
   - high priority memos (priority=1)
4. Update current state (project, phase, task, next_task)
5. Start phase if pending
6. Return task + manifest as JSON
```

### 8.3 Memo Scope Key Format

| Format | Scope | Example |
|--------|-------|---------|
| `key` | project | `notes` |
| `phase_id:key` | phase | `1:summary` |
| `phase_id:task_id:key` | task | `1:42:blockers` |

## 9. Dependencies

```
github.com/spf13/cobra v1.8.0      # CLI framework
github.com/mattn/go-sqlite3 v1.14.22  # SQLite driver
github.com/inconshreveable/mousetrap v1.1.0  # (indirect)
github.com/spf13/pflag v1.0.5      # (indirect)
```

## 10. Build & Install

```bash
# Build and install to $GOBIN or $GOPATH/bin
make install

# Uninstall
make uninstall
```

## 11. Recommendations

### Short-term
1. Add CLI command tests using `cobra.Command.Execute()` with mock DB
2. Add input validation for phase/task creation
3. Implement pagination for list commands

### Medium-term
1. Add `--format` flag for output (json/table/minimal)
2. Implement task dependencies (blocked_by field)
3. Add `clari undo` for recovering from failed tasks

### Long-term
1. Add plugin system for custom skills
2. Support remote DB (PostgreSQL) for team collaboration
3. Implement TTY handover for Claude Code integration

---

*Generated: 2026-02-03*
