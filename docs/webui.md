# Claribot Web UI Design Document

## 1. Overview

### 1.1 Purpose
A web dashboard for visually managing all Claribot features from the browser. Extends the operations previously only possible through Telegram and CLI with an intuitive UI.

### 1.2 Core Values
- **Visualization**: View Task trees and schedule timelines at a glance
- **Real-time**: Monitor Claude execution status and log streaming live
- **Convenience**: Manage Tasks with drag-and-drop, inline editing, one-click execution

### 1.3 Tech Stack
| Category | Choice | Reason |
|----------|--------|--------|
| **Framework** | React + TypeScript | Component-based, type safety |
| **Build** | Vite | Fast HMR, simple configuration |
| **UI Library** | shadcn/ui + Tailwind CSS | Customization freedom, lightweight |
| **State Management** | TanStack Query | Server state caching, auto-refresh |
| **Routing** | React Router v7 | SPA routing |
| **Icons** | Lucide React | Default icon set for shadcn/ui |
| **Deployment** | Go embed | Embed build output into claribot binary |

### 1.4 Directory Structure
```
claribot/
├── gui/                          # Web UI source code
│   ├── src/
│   │   ├── components/           # Shared UI components
│   │   │   ├── layout/           # Header, Sidebar, Layout
│   │   │   └── ui/               # shadcn/ui components
│   │   ├── pages/                # Page components
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Projects.tsx
│   │   │   ├── Tasks.tsx
│   │   │   ├── Messages.tsx
│   │   │   ├── Schedules.tsx
│   │   │   └── Settings.tsx
│   │   ├── hooks/                # Custom hooks
│   │   ├── api/                  # API client
│   │   ├── types/                # TypeScript types
│   │   └── App.tsx
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
├── bot/
│   └── internal/
│       └── webui/                # Go embed + HTTP handler
│           ├── webui.go          # embed.FS, static file serving
│           └── dist/             # Build output (gitignore)
```

---

## 2. API Integration Design

### 2.1 Using Existing API
Currently, claribot communicates through a single `POST /` endpoint with a JSON command structure. The web UI uses the same API.

```typescript
// api/client.ts
interface ClaribotRequest {
  command: string;
  subcommand: string;
  args: string[];
  flags: Record<string, any>;
}

interface ClaribotResponse {
  success: boolean;
  message: string;
  data: any;
  needs_input: boolean;
  prompt: string;
  context: string;
}

async function execute(req: ClaribotRequest): Promise<ClaribotResponse> {
  const res = await fetch('/api', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  return res.json();
}
```

### 2.2 Web UI-Specific Endpoints (To Be Added)

Additional minimal endpoints are needed for features not covered by the existing API.

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api` | POST | Existing command router proxy |
| `/api/stream` | WebSocket | Real-time Claude execution log streaming |
| `/api/health` | GET | Service health check (uptime, version) |

### 2.3 Real-time Communication
WebSocket is used to display progress in real time while Claude Code is running.

```
Client ──WebSocket──▶ /api/stream
                       │
                       ├─ claude:output   (Claude execution log)
                       ├─ task:updated    (Task status change)
                       ├─ message:updated (Message status change)
                       └─ schedule:fired  (Schedule execution notification)
```

---

## 3. Page Layout

### 3.1 Overall Layout

```
┌──────────────────────────────────────────────────┐
│  Header: Logo / Project Selector / System Status  │
├──────────┬───────────────────────────────────────┤
│          │                                       │
│ Sidebar  │           Main Content                │
│          │                                       │
│ Dashboard│                                       │
│ Projects │                                       │
│ Tasks    │                                       │
│ Messages │                                       │
│ Schedules│                                       │
│ Settings │                                       │
│          │                                       │
└──────────┴───────────────────────────────────────┘
```

**Header Components:**
- Left: Claribot logo
- Center: Project selection dropdown (shows currently selected project, includes "Global" option)
- Right: Claude status indicator (in use/idle/offline), system status badge

**Sidebar Components:**
- Navigation menu (icon + text)
- Collapse mode support (icon only)
- Active menu highlight

---

### 3.2 Dashboard

**Path**: `/`

Main page for an at-a-glance overview of the entire system.

```
┌─────────────────────────────────────────────────┐
│                    Dashboard                     │
├──────────┬──────────┬──────────┬────────────────┤
│ Claude   │ Tasks    │ Messages │ Schedules      │
│ ● 2/3    │ 12 total │ 3 active │ 5 active       │
│ in use   │ 4 pending│ 47 done  │ next: 07:00    │
└──────────┴──────────┴──────────┴────────────────┘
│                                                  │
│  [Recent Activity Timeline]                      │
│  ─ 14:23 Task #7 "Implement API endpoints" done  │
│  ─ 14:20 Schedule #3 execution started           │
│  ─ 14:15 Message "Review the code" processed     │
│  ─ 14:00 Task #5 "DB schema design" plan created │
│                                                  │
├──────────────────┬───────────────────────────────┤
│ Task Status Dist.│  Tasks by Project              │
│ [Pie Chart]      │  [Bar Chart]                   │
│ todo: 4          │  claribot: 8                   │
│ planned: 3       │  blog: 3                       │
│ done: 12         │  api-server: 1                 │
│ failed: 1        │                               │
└──────────────────┴───────────────────────────────┘
```

**Components:**
1. **4 Summary Cards**: Claude usage, Task status, Message status, Schedule status
2. **Recent Activity Timeline**: Recent events across the system (Task completions, message processing, schedule executions, etc.)
3. **Task Status Distribution Chart**: Status ratio of Tasks in the current project
4. **Activity by Project**: Active Task count per project

**Data Refresh**: Auto-polling every 30 seconds via TanStack Query + immediate refresh on WebSocket events

---

### 3.3 Project Management (Projects)

**Path**: `/projects`

```
┌─────────────────────────────────────────────────┐
│  Projects                      [+ Add Project]   │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌─────────────────────────────────────────┐    │
│  │ claribot                                │    │
│  │ ~/projects/claribot                     │    │
│  │ LLM-based project automation system     │    │
│  │ Tasks: 12 | Active Schedules: 3         │    │
│  │                   [Select] [Edit] [Delete]│   │
│  └─────────────────────────────────────────┘    │
│                                                  │
│  ┌─────────────────────────────────────────┐    │
│  │ blog                                    │    │
│  │ ~/projects/blog                         │    │
│  │ Personal blog                           │    │
│  │ Tasks: 3 | Active Schedules: 1          │    │
│  └─────────────────────────────────────────┘    │
│                                                  │
└─────────────────────────────────────────────────┘
```

**Features:**
- Project card list (grid/list view toggle)
- Cards display project ID, path, description, Task/Schedule summary
- `[Select]` button: Switch to that project (synced with Header dropdown)
- `[+ Add Project]`: Modal dialog (path, description input)

---

### 3.4 Task Management (Tasks) - Core Page

**Path**: `/tasks`

This is the core page of the Claribot web UI. It visually manages the Task tree structure.

#### 3.4.1 Task Tree View (Default View)

```
┌────────────────────────────────────────────────────────────┐
│  Tasks (claribot)           [Tree] [List]                  │
│                            [+ Add Task] [▶ Plan All] [▶ Run All] [Cycle All] │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ▼ #1 API Server Refactoring          split         depth:0│
│  │                                                         │
│  ├─ ▼ #2 Router Modularization        split         depth:1│
│  │  ├── #4 Handler Separation         ● planned     depth:2│
│  │  └── #5 Add Middleware             ● planned     depth:2│
│  │                                                         │
│  ├── #3 Error Handling Unification    ✅ done        depth:1│
│  │                                                         │
│  └── #6 Write Tests                  ○ todo        depth:1 │
│                                                            │
│  #7 Documentation Update              ○ todo        depth:0│
│                                                            │
└────────────────────────────────────────────────────────────┘
```

**Tree Interactions:**
- Click: Open Task detail panel (right slide)
- Double-click: Inline title editing
- Drag: Reorder Tasks (within same parent)
- Expand/collapse nodes

**Status Icons:**
- `○` todo (gray)
- `◐` split (blue)
- `●` planned (yellow)
- `✅` done (green)
- `❌` failed (red)

#### 3.4.2 List View

```
┌──────────────────────────────────────────────────────────┐
│  [Tree] [List]                                           │
│  Filter: [All ▼] [Status: All ▼] [Leaf only □]          │
├────┬───────────────────┬────────────┬───────┬────────────┤
│ ID │ Title             │ Status     │ Depth │ Parent     │
├────┼───────────────────┼────────────┼───────┼────────────┤
│ 1  │ API Server Refact.│ split      │ 0     │ -          │
│ 2  │ Router Modulariz. │ split      │ 1     │ #1         │
│ 3  │ Error Handling    │ done       │ 1     │ #1         │
│ 4  │ Handler Separation│ planned    │ 2     │ #2         │
│ 5  │ Add Middleware    │ planned    │ 2     │ #2         │
│ 6  │ Write Tests       │ todo       │ 1     │ #1         │
│ 7  │ Doc Update        │ todo       │ 0     │ -          │
└────┴───────────────────┴────────────┴───────┴────────────┘
```

**Filtering:**
- By status: todo, split, planned, done, failed
- Leaf only view (is_leaf = true)
- Filter by parent Task

#### 3.4.3 Task Detail Panel (Right Slide)

```
┌───────────────────────────────┐
│  Task #4                  [×] │
│  Handler Separation           │
├───────────────────────────────┤
│                               │
│  Status: ● planned            │
│  Parent: #2 Router Modular.   │
│  Depth: 2                     │
│  Leaf: Yes                    │
│  Created: 2025-01-15 14:00    │
│  Updated: 2025-01-15 15:30    │
│                               │
│  ── Spec ──────────────────── │
│  Separate HTTP handlers by    │
│  function into individual     │
│  files.                       │
│  - project handler            │
│  - task handler               │
│  - message handler            │
│  [Edit]                       │
│                               │
│  ── Plan ──────────────────── │
│  ## Implementation Approach   │
│  1. Extract each command      │
│     handler from              │
│     handler/router.go         │
│  2. Create handler/project.go │
│  3. Create handler/task.go    │
│  [Edit]                       │
│                               │
│  ── Report ────────────────── │
│  (Not yet executed)           │
│                               │
│  ─────────────────────────── │
│  [▶ Plan] [▶ Run] [Delete]   │
│                               │
└───────────────────────────────┘
```

**Features:**
- Render Spec, Plan, Report as markdown
- Inline editing for each field (markdown editor)
- Status change dropdown
- Individual execution via Plan/Run buttons

#### 3.4.4 Task Action Buttons

| Button | Action | API Call |
|--------|--------|---------|
| `+ Add Task` | Create new Task modal (title, parent, Spec input) | `task add` |
| `▶ Plan All` | Generate plans for all todo Tasks | `task plan --all` |
| `▶ Run All` | Execute all planned leaf Tasks | `task run --all` |
| `Cycle All` | Auto-cycle Plan + Run | `task cycle` |
| `▶ Plan` (individual) | Generate plan for single Task | `task plan <id>` |
| `▶ Run` (individual) | Execute single Task | `task run <id>` |

---

### 3.5 Messages

**Path**: `/messages`

```
┌─────────────────────────────────────────────────────────┐
│  Messages (claribot)                                     │
│  [Status Filter: All ▼]                                  │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌─ Message Input ──────────────────────────────────┐   │
│  │                                                  │   │
│  │  Enter a message to send to Claude...            │   │
│  │                                          [Send]  │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  ── Message History ────────────────────────────────── │
│                                                          │
│  #47  "Organize the API endpoint list"                   │
│  cli | ✅ done | 2025-01-15 14:23                        │
│  ▼ View Result                                           │
│  ┌──────────────────────────────────────────────────┐   │
│  │ ## API Endpoint List                              │   │
│  │ 1. POST /api - Execute command                    │   │
│  │ 2. GET /api/health - Health check                 │   │
│  │ ...                                               │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  #46  "Review the code"                                  │
│  telegram | processing | 2025-01-15 14:20                │
│  ▼ Live Log                                              │
│  ┌──────────────────────────────────────────────────┐   │
│  │ [Real-time Claude output streaming...]            │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  #45  "Run the tests"                                    │
│  schedule | ✅ done | 2025-01-15 14:00                   │
│                                                          │
│  [Load More]                                             │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

**Features:**
- Top message input: Send messages directly to Claude (`message send cli <text>`)
- Message history: Sorted by newest, infinite scroll
- Status filter: pending, processing, done, failed
- Source icon: telegram, cli, schedule
- Expand/collapse results (markdown rendering)
- Processing status messages: Real-time log streaming via WebSocket

---

### 3.6 Schedules

**Path**: `/schedules`

```
┌─────────────────────────────────────────────────────────┐
│  Schedules                          [+ Add Schedule]     │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌───────────────────────────────────────────────────┐  │
│  │ #1  Daily Code Review                       [ON]  │  │
│  │ 0 9 * * 1-5  (Weekdays 09:00)                    │  │
│  │ claribot                                          │  │
│  │ "Review code changes and organize issues"         │  │
│  │ Last run: 2025-01-15 09:00 ✅                     │  │
│  │ Next run: 2025-01-16 09:00                        │  │
│  │                    [Execution Log] [Edit] [Delete] │  │
│  └───────────────────────────────────────────────────┘  │
│                                                          │
│  ┌───────────────────────────────────────────────────┐  │
│  │ #2  Daily Report                            [ON]  │  │
│  │ 0 18 * * *  (Daily 18:00)                         │  │
│  │ (Global)                                          │  │
│  │ "Write a summary report of today's work"          │  │
│  │ Last run: 2025-01-15 18:00 ✅                     │  │
│  │ Next run: 2025-01-16 18:00                        │  │
│  └───────────────────────────────────────────────────┘  │
│                                                          │
│  ┌───────────────────────────────────────────────────┐  │
│  │ #3  One-time Notification                   [OFF] │  │
│  │ 30 14 * * *  (Single execution completed)         │  │
│  │ (Global)                                          │  │
│  │ "Notification test"                               │  │
│  │ run_once: true | Disabled                         │  │
│  └───────────────────────────────────────────────────┘  │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

**Add Schedule Modal:**
```
┌─────────────────────────────────┐
│  Add New Schedule           [×] │
├─────────────────────────────────┤
│                                 │
│  Cron Expression:               │
│  ┌─────────────────────────┐   │
│  │ 0 9 * * 1-5             │   │
│  └─────────────────────────┘   │
│  → Weekdays 09:00 (readable)   │
│                                 │
│  Message:                       │
│  ┌─────────────────────────┐   │
│  │ Review the code          │   │
│  └─────────────────────────┘   │
│                                 │
│  Project:                       │
│  [claribot ▼]                   │
│                                 │
│  ☐ Run once (run_once)          │
│                                 │
│            [Cancel] [Add]       │
└─────────────────────────────────┘
```

**Execution Log View (Modal or Sub-page):**
```
┌─────────────────────────────────────────────┐
│  Schedule #1 Execution Log              [×] │
├──────┬──────────┬────────┬──────────────────┤
│ Run  │ Start    │ Status │ Duration         │
├──────┼──────────┼────────┼──────────────────┤
│ #15  │ 01-15 09:00 │ ✅ done │ 2m 34s      │
│ #14  │ 01-14 09:00 │ ✅ done │ 1m 50s      │
│ #13  │ 01-13 09:00 │ ❌ failed│ 0m 12s     │
│ #12  │ 01-10 09:00 │ ✅ done │ 3m 10s      │
└──────┴──────────┴────────┴──────────────────┘
│                                              │
│ Run #15 Detail:                              │
│ ┌──────────────────────────────────────┐    │
│ │ ## Code Review Results                │    │
│ │ 1. handler/router.go - Improve error │    │
│ │    handling                           │    │
│ │ 2. task/plan.go - Add comments       │    │
│ │ ...                                   │    │
│ └──────────────────────────────────────┘    │
└──────────────────────────────────────────────┘
```

**Features:**
- Schedule card list
- ON/OFF toggle switch (`schedule enable/disable`)
- Convert cron expressions to human-readable format for display
- View execution history and results
- Filter by project

---

### 3.7 Settings

**Path**: `/settings`

```
┌─────────────────────────────────────────────────┐
│  Settings                                        │
├─────────────────────────────────────────────────┤
│                                                  │
│  ── System Info ─────────────────────────────── │
│  Claribot Version: v0.2.20                       │
│  CLI Version: v0.2.5                             │
│  Uptime: 3d 14h 22m                             │
│  DB Path: ~/.claribot/db.clt                     │
│                                                  │
│  ── Claude Code ─────────────────────────────── │
│  Max Concurrent Executions: 3                    │
│  Currently In Use: 1                             │
│  Idle Timeout: 1200s                             │
│                                                  │
│  ── Active Sessions ─────────────────────────── │
│  #1: claribot (task plan) - 2m 30s               │
│  #2: (idle)                                      │
│  #3: (idle)                                      │
│                                                  │
│  ── Connection Status ───────────────────────── │
│  Service: ● Connected (127.0.0.1:9847)           │
│  Telegram: ● Connected                           │
│  WebSocket: ● Connected                          │
│                                                  │
└─────────────────────────────────────────────────┘
```

---

## 4. Common UI Components

### 4.1 Claude Execution Modal

Modal/toast displayed while Claude Code is running:

```
┌─────────────────────────────────────┐
│  Claude Running...              [×] │
│                                     │
│  Running Task #4 "Handler Separation"│
│  Elapsed: 1m 24s                    │
│                                     │
│  ┌─────────────────────────────┐   │
│  │ > Analyzing handler/router.go│   │
│  │ > Separating project handler │   │
│  │ > Creating handler/project.go│   │
│  │ > ...                       │   │
│  │ [Real-time log streaming]   │   │
│  └─────────────────────────────┘   │
│                                     │
└─────────────────────────────────────┘
```

### 4.2 Confirmation Dialog

Displayed for dangerous operations like deletion:

```
┌──────────────────────────────┐
│  Confirm Deletion            │
│                              │
│  Delete Task #4              │
│  "Handler Separation"?       │
│                              │
│  This action cannot be undone.│
│                              │
│         [Cancel] [Delete]    │
└──────────────────────────────┘
```

### 4.3 Toast Notifications

```
┌──────────────────────────────┐
│ ✅ Task #4 plan created       │  ← Success (green, auto-close after 3s)
└──────────────────────────────┘

┌──────────────────────────────┐
│ ❌ Schedule add failed:       │  ← Error (red, manual close)
│    Invalid cron expression   │
└──────────────────────────────┘

┌──────────────────────────────┐
│ Schedule #1 execution done   │  ← WebSocket notification (blue)
└──────────────────────────────┘
```

---

## 5. Deployment & Integration

### 5.1 Go embed Integration

The web UI build output is embedded into the Go binary, requiring no separate web server.

```go
// bot/internal/webui/webui.go
package webui

import "embed"

//go:embed dist/*
var staticFiles embed.FS
```

**Build Flow:**
```
gui/ → npm run build → gui/dist/ → cp → bot/internal/webui/dist/ → go build (embed) → claribot binary
```

### 5.2 HTTP Routing

```go
// Added to existing claribot HTTP server
mux.HandleFunc("/api", handleAPI)          // Existing command router
mux.HandleFunc("/api/stream", handleWS)    // WebSocket
mux.HandleFunc("/api/health", handleHealth)
mux.Handle("/", webui.Handler())           // Static file serving (SPA fallback)
```

**SPA Fallback**: All requests not starting with `/api` are redirected to `index.html` (React Router support).

### 5.3 Makefile Additions

```makefile
build-gui:
	cd gui && npm install && npm run build
	rm -rf bot/internal/webui/dist
	cp -r gui/dist bot/internal/webui/dist

build: build-gui build-cli build-bot

dev-gui:
	cd gui && npm run dev
```

### 5.4 Configuration

```yaml
# config.yaml addition
webui:
  enabled: true       # Enable/disable web UI
  port: 9848          # Web UI port (separate) or share existing port
```

---

## 6. Implementation Priority

### Phase 1: Foundation (MVP) ✅
1. ~~Project scaffolding (Vite + React + TypeScript + shadcn/ui)~~ ✅
2. ~~API client module (`/api` POST communication)~~ ✅
3. ~~Layout component (Header + Sidebar + Main)~~ ✅
4. ~~Project selector (dropdown)~~ ✅
5. ~~Dashboard page (4 summary cards)~~ ✅
6. ~~Go embed integration and static file serving~~ ✅

### Phase 2: Core Features ✅
7. ~~Projects page (CRUD)~~ ✅
8. ~~Tasks page - List view~~ ✅
9. ~~Tasks page - Tree view~~ ✅
10. ~~Task detail panel (Spec/Plan/Report display and editing)~~ ✅
11. ~~Task execution buttons (Plan/Run/Cycle)~~ ✅
12. ~~Messages page (history + send)~~ ✅

### Phase 3: Visualization
13. ~~Dashboard charts (Task status distribution)~~ ✅

### Phase 4: Real-time
16. WebSocket integration (`/api/stream`)
17. Real-time Claude execution log streaming
18. Real-time Task/Message status updates
19. Toast notifications (schedule execution completed, etc.)

### Phase 5: Polish ✅ (Partial)
20. ~~Schedules page (CRUD + execution history)~~ ✅
21. ~~Settings page~~ ✅
22. Dark mode support (CSS variables ready, toggle not implemented)
23. Responsive layout (mobile support)
24. Keyboard shortcuts (Task navigation, execution)

---

## 7. Screen Flow Diagram

```
[Dashboard] ──── Select Project ────▶ [Dashboard Refresh]
     │
     ├──▶ [Projects] ── Select Project ──▶ [Tasks]
     │
     ├──▶ [Tasks]
     │      ├── Tree View ── Click Task ──▶ [Task Detail Panel]
     │      ├── List View ── Click Row ──▶ [Task Detail Panel]
     │      └── Plan/Run/Cycle Button ──▶ [Claude Execution Modal]
     │
     ├──▶ [Messages]
     │      ├── Send Message ──▶ [Claude Execution Modal]
     │      └── Expand/Collapse Result
     │
     ├──▶ [Schedules]
     │      ├── Add Schedule ──▶ [Add Modal]
     │      ├── ON/OFF Toggle
     │      └── Execution Log ──▶ [Execution Log View]
     │
     └──▶ [Settings]
```
