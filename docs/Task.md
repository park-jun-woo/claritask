# Task System Design

> Claribot's Task-based divide-and-conquer system

---

## Overview

Recursively splits large tasks and injects context from related Tasks to ensure consistent execution.

**Core Idea**: In the 1st pass, Claude decides whether to split or plan, traversing down to leaf nodes via DFS.

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

**Order**: Deepest first (depth DESC)

```sql
SELECT * FROM tasks
WHERE is_leaf = 1 AND status = 'planned'
ORDER BY depth DESC, id ASC
```

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
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_tasks_leaf ON tasks(is_leaf);
```

---

## System Prompt

Prompt injected to Claude during the 1st pass:

`bot/internal/prompts/dev.platform/task.md`

---

## Implementation Status

### Phase 1-5: Existing Features ✅

### Phase 6: Recursive Splitting System ✅
- [x] Added `IsLeaf`, `Depth` to Task struct
- [x] Added `is_leaf`, `depth`, `split` status to DB schema
- [x] Added `task add --spec` flag
- [x] 1st pass prompt template (`task.md`)
- [x] Output parser (`[SPLIT]`, `[PLANNED]`)
- [x] `planRecursive()` recursive traversal
- [x] Added leaf condition for 2nd pass

---

*Claribot Task System v0.3*
