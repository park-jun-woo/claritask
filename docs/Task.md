# Task System Design

> Claribot's Task-based divide-and-conquer system

---

## Overview

Recursively splits large tasks and injects context from related Tasks to ensure consistent execution.

**Core Idea**: In the 1st pass, Claude decides whether to split or plan, traversing down to leaf nodes via DFS. In the 2nd pass, planned leaf tasks are executed in priority/depth order with optional parallel execution.

---

## Task Structure

```go
Task {
    ID          int
    ParentID    *int      // Parent Task
    Title       string    // Title
    Spec        string    // Requirement specification (immutable)
    Plan        string    // Plan (leaf only)
    Report      string    // Completion report (after 2nd pass)
    Status      string    // todo → split/planned → done
    Error       string
    IsLeaf      bool      // true: execution target, false: subdivided
    Depth       int       // Tree depth (root=0)
    Priority    int       // Execution priority (higher = executed first)
    CreatedAt   string
    UpdatedAt   string
}
```

### State Transitions

```
todo ─┬─→ split (subdivided, is_leaf=false)
      │
      └─→ planned (plan ready, is_leaf=true) → done
```

| Status | Description | is_leaf |
|--------|-------------|---------|
| todo | Registered, not yet processed | true (default) |
| split | Subdivided, has child Tasks | false |
| planned | Planning complete, awaiting execution | true |
| done | Execution complete | true |
| failed | Failed | - |

---

## Traversal System

Traversals track the execution of plan/run/cycle operations. Each traversal records its type, status, and result counts.

### Traversal Types

| Type | Description |
|------|-------------|
| plan | 1st pass - recursive splitting of todo tasks |
| run | 2nd pass - execution of planned leaf tasks |
| cycle | Full cycle - plan phase followed by run phase |

### Traversal Status

```
running → done      (completed successfully)
        → failed    (error occurred)
        → cancelled (stopped by user via task stop)
```

### Cycle Flow

The `task cycle` command orchestrates both phases automatically:

```
Cycle:
  PHASE 1: PLAN ALL (repeating, max 10 iterations)
    Loop:
      - Count todo tasks
      - If zero → exit loop
      - PlanAll: plan all todo tasks
      - Newly split tasks create more todo tasks
      - Repeat until no todo tasks remain

  PHASE 2: RUN ALL
    - Check cancel flag
    - Count planned tasks
    - If > 0 → RunAll: execute all planned leaf tasks
    - Tasks become done or failed

  SUMMARY:
    - Output done/failed counts
```

---

## 1st Pass (Recursive Splitting)

### Flow

```
Task A (todo)
├─ Claude decides: needs splitting
├─ Creates child Tasks B, C via clari task add
├─ Task A → split
├─ Immediately traverse Task B
│   └─ Claude decides: can plan → planned
└─ Immediately traverse Task C
    ├─ Claude decides: needs splitting
    ├─ Creates Tasks D, E
    ├─ Task C → split
    ├─ Traverse Task D → planned
    └─ Traverse Task E → planned
```

### Ordering

```sql
SELECT * FROM tasks
WHERE is_leaf = 1 AND status = 'todo'
ORDER BY priority DESC, depth ASC, id ASC
```

### Claude Decision Criteria

**Split conditions** (if any apply):
- 3 or more independent steps
- Multiple domains (UI + API + DB)
- Expected file changes exceed 5

**Plan conditions** (all must be met):
- Single purpose
- 5 or fewer file changes
- Can be executed independently

### Output Format

Claude responds with `[SPLIT]` or `[PLANNED]` markers. `parser.go` handles code block stripping and marker search with fallback.

```go
PlanResult {
    Type     string    // "split" or "planned"
    Plan     string    // Plan content (if planned)
    Children []Child   // Child tasks (if split)
}
```

```
[SPLIT]
- Task #<id>: <title>
- Task #<id>: <title>
```

or

```
[PLANNED]
## Implementation Approach
...
```

### Depth Limit

Force plan creation when `MaxDepth = 5` is reached.

---

## 2nd Pass (Execution)

**Target**: `is_leaf = true AND status = 'planned'`

**Order**: Priority first, then deepest first

```sql
SELECT * FROM tasks
WHERE is_leaf = 1 AND status = 'planned'
ORDER BY priority DESC, depth DESC, id ASC
```

### Parallel Execution

Tasks can be executed concurrently based on the project's `parallel` config setting.

- **Default**: 3 concurrent workers
- **Config**: Stored in project's local DB `config` table (`key = 'parallel'`)
- **Sequential mode**: Set parallel to 1

Implementation uses a semaphore pattern with buffered channels. Each worker gets its own DB connection for SQLite concurrency safety. Workers check the cancel flag before processing each task.

---

## Cycle State Tracking

Per-project cycle state is managed in memory (`cycle_state.go`) to monitor ongoing traversals.

```go
CycleState {
    Running       bool       // Whether cycle is active
    Type          string     // "cycle", "plan", "run"
    StartedAt     time.Time
    CurrentTaskID int        // Currently processing task
    ProjectPath   string
    ProjectID     string
    ActiveWorkers int        // Number of parallel workers running
    Phase         string     // Current phase: "plan" or "run"
    TargetTotal   int        // Total tasks in current phase
    Completed     int        // Tasks completed in current phase
}
```

Key functions: `SetCycleState`, `GetCycleState`, `GetAllCycleStates`, `ClearCycleState`, `UpdateCurrentTask`, `UpdatePhase`, `IncrementCompleted`, `UpdateActiveWorkers`, `GetActiveWorkers`, `ResetActiveWorkers`, `IsCycleRunning`, `IsAnyCycleRunning`, `SetCycleCancel`, `CancelCycle`, `CancelAllCycles`.

### CycleStatusInfo

`GetCycleStatus()` and `GetAllCycleStatuses()` return display-ready status info combining CycleState with Claude process status:

```go
CycleStatusInfo {
    Status        string    // "running", "interrupted", "idle"
    Type          string    // "cycle", "plan", "run"
    StartedAt     time.Time
    CurrentTaskID int
    ProjectPath   string
    ProjectID     string
    ActiveWorkers int
    Phase         string    // "plan", "run"
    TargetTotal   int
    Completed     int
}
```

- `running`: Claude processes are active
- `interrupted`: CycleState exists but no Claude processes found
- `idle`: No active cycles

### Graceful Stop

`state.go` provides atomic cancel flag and stop functions:

- `RequestCancel()` / `IsCancelled()` / `ResetCancel()` - global atomic bool
- `Stop()` - cancels all running cycles across all projects
- `StopProject(projectPath)` - cancels a specific project's cycle

Workers check `IsCancelled()` and `ctx.Err()` before processing each task.

### Auth Error Handling

During traversals, if Claude returns an authentication error (detected by `claude.IsAuthError()`), the current worker aborts and propagates `ErrorType: "auth_error"` upward. In parallel mode, the parent context is cancelled, stopping all remaining workers. This prevents wasting API quota on repeated auth failures.

---

## Task Statistics

```go
Stats {
    Total      int    // Total task count
    Leaf       int    // Execution targets (is_leaf=1)
    Todo       int    // Awaiting plan
    Planned    int    // Plan ready, awaiting execution
    Done       int    // Completed
    Failed     int    // Failed
    InProgress int    // Currently executing (runtime, not from DB)
}
```

> `InProgress` is not stored in the DB. It is tracked at runtime via CycleState's active workers and populated by the API layer when returning stats.

---

## Context Map

When executing tasks (both plan and run phases), a Context Map is injected into the prompt. It provides a lightweight summary of the entire task tree so Claude understands the broader context.

### Format

```
#1 [todo] Main Task
  #2 [planned] Subtask A
  #3 [split] Subtask B
    #4 [todo] Sub-subtask
```

Built by `BuildContextMap()` which queries all tasks with their status and depth, formatted with indentation by depth level.

---

## CLI Commands

### task add

```bash
# Basic
clari task add "title"

# Specify parent
clari task add "title" --parent 1

# Include Spec (used during Claude splitting)
clari task add "title" --parent 1 --spec "requirements"

# Load Spec from file
clari task add "title" --spec-file path/to/spec.md
```

### task get

```bash
# Get task details (spec, plan, report)
clari task get <id>
```

### task list

```bash
# List tasks (optionally filter by parent)
clari task list [parent_id]

# Tree view
clari task list --tree
```

### task set

```bash
# Update task field
clari task set <id> <field> <value>

# Supported fields: title, spec, plan, report, status, priority
clari task set <id> priority 10
```

### task plan

```bash
# Single Task (recursive splitting)
clari task plan [id]

# All todo Tasks
clari task plan --all
```

### task run

```bash
# Execute single leaf Task
clari task run [id]

# Execute all planned leaf Tasks
clari task run --all
```

### task cycle

```bash
# Full cycle: repeat PlanAll until no todo tasks, then RunAll
clari task cycle
```

### task stop

```bash
# Stop all running cycles (graceful cancellation)
clari task stop
```

### task delete

```bash
# Delete a task
clari task delete <id>
```

---

## DB Schema

```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER,
    title TEXT NOT NULL,
    spec TEXT DEFAULT '',
    plan TEXT DEFAULT '',
    report TEXT DEFAULT '',
    status TEXT DEFAULT 'todo'
        CHECK(status IN ('todo', 'split', 'planned', 'done', 'failed')),
    error TEXT DEFAULT '',
    is_leaf INTEGER DEFAULT 1,
    depth INTEGER DEFAULT 0,
    priority INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_tasks_parent ON tasks(parent_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_leaf ON tasks(is_leaf);

CREATE TABLE traversals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK(type IN ('plan', 'run', 'cycle')),
    target_id INTEGER,
    trigger TEXT DEFAULT '',
    status TEXT DEFAULT 'running'
        CHECK(status IN ('running', 'done', 'failed', 'cancelled')),
    total INTEGER DEFAULT 0,
    success INTEGER DEFAULT 0,
    failed INTEGER DEFAULT 0,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    FOREIGN KEY (target_id) REFERENCES tasks(id) ON DELETE SET NULL
);

CREATE INDEX idx_traversals_type ON traversals(type);
CREATE INDEX idx_traversals_status ON traversals(status);

CREATE TABLE config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE specs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT DEFAULT '',
    status TEXT DEFAULT 'draft'
        CHECK(status IN ('draft', 'review', 'approved', 'deprecated')),
    priority INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_specs_status ON specs(status);
```

---

## Prompt Templates

Prompts are built in `bot/internal/task/prompt.go` using Go `text/template` with fallback to hardcoded simple prompts.

| Phase | Template | Data Struct |
|-------|----------|-------------|
| 1st pass (Plan) | `bot/internal/prompts/common/task.md` | `PlanPromptData` |
| 2nd pass (Run) | `bot/internal/prompts/common/task_run.md` | `ExecutePromptData` |

Both prompts include the Context Map and a report file path. The report file is written by Claude during execution, then read and stored in DB, and the file is deleted.

---

## Source Files

| File | Description |
|------|-------------|
| `task.go` | Task/Stats structs, MaxDepth constant, notifier init, formatDuration helper |
| `add.go` | Task creation with parent/depth calculation |
| `get.go` | Task detail retrieval |
| `list.go` | List (paginated), ListTree (full tree), statusToIcon helper |
| `set.go` | Field update with validation |
| `delete.go` | Task deletion with confirmation |
| `stats.go` | GetStats - task statistics query (COALESCE for NULL safety) |
| `plan.go` | Plan/PlanAll - 1st pass recursive splitting (sequential + parallel) |
| `run.go` | Run/RunAll/RunWithContext - 2nd pass execution (sequential + parallel), getParallel config |
| `cycle.go` | Cycle - full plan+run orchestration (maxCycleIterations=10) |
| `cycle_state.go` | Per-project cycle state, CycleStatusInfo, GetCycleStatus/GetAllCycleStatuses |
| `state.go` | Cancel flag and Stop/StopProject |
| `prompt.go` | Prompt template building (PlanPromptData, ExecutePromptData) with fallback |
| `parser.go` | PlanResult/Child structs, ParsePlanOutput, stripCodeBlocks, extractMarker |
| `context_map.go` | BuildContextMap - task tree summary |
| `traversal.go` | Traversal DB insert/finish, countFromMessage regex parser |
| `related.go` | GetRelated - parent/child task lookup |

---

*Claribot Task System v0.7*
