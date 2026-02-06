# Claribot Web UI Design Document

## 1. Overview

### 1.1 Purpose
A web dashboard for visually managing all Claribot features from the browser. Extends the operations previously only possible through Telegram and CLI with an intuitive UI.

### 1.2 Core Values
- **Visualization**: View Task trees, project stats boards, and cycle progress at a glance
- **Real-time**: Monitor Claude execution status with auto-polling and cycle progress tracking
- **Convenience**: Manage Tasks with inline editing, one-click execution, and chat-style messaging

### 1.3 Tech Stack
| Category | Choice | Reason |
|----------|--------|--------|
| **Framework** | React + TypeScript | Component-based, type safety |
| **Build** | Vite | Fast HMR, simple configuration |
| **UI Library** | shadcn/ui + Tailwind CSS | Customization freedom, lightweight |
| **State Management** | TanStack Query | Server state caching, auto-refresh |
| **Routing** | React Router v7 | SPA routing |
| **Icons** | Lucide React | Default icon set for shadcn/ui |
| **Markdown** | react-markdown + remark-gfm | Spec/Plan/Report HTML rendering (MarkdownRenderer component) |
| **QR Code** | qrcode.react (QRCodeSVG) | TOTP setup QR generation |
| **YAML** | yaml (npm) | Config YAML parsing/serialization in Settings page |
| **Deployment** | Go embed | Embed build output into claribot binary |

### 1.4 Directory Structure
```
claribot/
├── gui/                          # Web UI source code
│   ├── src/
│   │   ├── components/           # Shared UI components
│   │   │   ├── layout/           # Header, Sidebar, Layout
│   │   │   ├── ui/               # shadcn/ui components
│   │   │   ├── ProjectSelector.tsx  # Project dropdown selector
│   │   │   ├── ChatBubble.tsx       # Chat message bubble component
│   │   │   └── MarkdownRenderer.tsx # Markdown-to-HTML renderer
│   │   ├── pages/                # Page components
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Projects.tsx
│   │   │   ├── ProjectEdit.tsx
│   │   │   ├── Tasks.tsx
│   │   │   ├── Messages.tsx
│   │   │   ├── Schedules.tsx
│   │   │   ├── Specs.tsx
│   │   │   ├── Settings.tsx
│   │   │   ├── Login.tsx
│   │   │   └── Setup.tsx
│   │   ├── hooks/                # Custom hooks
│   │   │   ├── useClaribot.ts    # TanStack Query hooks for all APIs
│   │   │   └── useAuth.ts        # Authentication hooks (login, logout, setup)
│   │   ├── api/                  # API client
│   │   │   └── client.ts         # RESTful API client (apiGet/apiPost/apiPatch/apiPut/apiDelete)
│   │   ├── types/                # TypeScript types
│   │   │   └── index.ts          # All type definitions
│   │   └── App.tsx               # Routing + Auth guard
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
├── bot/
│   └── internal/
│       ├── handler/
│       │   └── restful.go        # RESTful API router + handlers
│       └── webui/                # Go embed + HTTP handler
│           ├── webui.go          # embed.FS, static file serving
│           └── dist/             # Build output (gitignore)
```

---

## 2. API Integration Design

### 2.1 RESTful API

The web UI communicates with the claribot backend via RESTful API endpoints. All endpoints are prefixed with `/api/`.

```typescript
// api/client.ts - Separate HTTP helpers with credential support
async function apiGet<T>(path: string): Promise<T> {
  const res = await fetch(`/api${path}`, { credentials: 'include' });
  if (!res.ok) throw new Error(`API error: ${res.status} ${res.statusText}`);
  return res.json();
}

async function apiPost<T>(path: string, body?: unknown): Promise<T> { ... }
async function apiPatch<T>(path: string, body: unknown): Promise<T> { ... }
async function apiPut<T>(path: string, body: unknown): Promise<T> { ... }
async function apiDelete<T>(path: string): Promise<T> { ... }
```

### 2.2 API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/health` | GET | Service health check (version, uptime) |
| `/api/status` | GET | Claude status + cycle status + task stats |
| `/api/usage` | GET | Claude Code usage statistics (stats-cache.json based) |
| `/api/usage/refresh` | POST | Refresh live usage data from PTY |
| `/api/auth/setup` | POST | Initial password setup + TOTP verification (combined) |
| `/api/auth/totp-setup` | GET | Get TOTP URI for QR code generation |
| `/api/auth/login` | POST | Login with password + TOTP |
| `/api/auth/logout` | POST | Logout (clear session) |
| `/api/auth/status` | GET | Check authentication status |
| `/api/projects` | GET | List all projects |
| `/api/projects` | POST | Add/create project |
| `/api/projects/stats` | GET | All project task statistics |
| `/api/projects/{id}` | GET | Get project details |
| `/api/projects/{id}` | PATCH | Update project settings |
| `/api/projects/{id}` | DELETE | Delete project |
| `/api/projects/{id}/switch` | POST | Switch active project |
| `/api/tasks` | GET | List tasks (supports `?tree=true`, `?parent_id=`, `?all=true`) |
| `/api/tasks` | POST | Add new task |
| `/api/tasks/plan-all` | POST | Plan all todo tasks |
| `/api/tasks/run-all` | POST | Run all planned leaf tasks |
| `/api/tasks/cycle` | POST | Cycle all tasks (Plan + Run) |
| `/api/tasks/stop` | POST | Stop active cycle |
| `/api/tasks/{id}` | GET | Get task details |
| `/api/tasks/{id}` | PATCH | Update task fields |
| `/api/tasks/{id}` | DELETE | Delete task |
| `/api/tasks/{id}/plan` | POST | Plan single task |
| `/api/tasks/{id}/run` | POST | Run single task |
| `/api/messages` | GET | List messages |
| `/api/messages` | POST | Send new message |
| `/api/messages/status` | GET | Message status summary |
| `/api/messages/processing` | GET | Currently processing messages |
| `/api/messages/{id}` | GET | Get message details |
| `/api/schedules` | GET | List schedules |
| `/api/schedules` | POST | Add schedule |
| `/api/schedules/{id}` | GET | Get schedule details |
| `/api/schedules/{id}` | PATCH | Update schedule fields |
| `/api/schedules/{id}` | DELETE | Delete schedule |
| `/api/schedules/{id}/enable` | POST | Enable schedule |
| `/api/schedules/{id}/disable` | POST | Disable schedule |
| `/api/schedules/{id}/runs` | GET | Get schedule run history |
| `/api/schedule-runs/{runId}` | GET | Get individual run detail |
| `/api/specs` | GET | List specs |
| `/api/specs` | POST | Add spec |
| `/api/specs/{id}` | GET | Get spec details |
| `/api/specs/{id}` | PATCH | Update spec |
| `/api/specs/{id}` | DELETE | Delete spec |
| `/api/configs` | GET | List all config key-value pairs |
| `/api/configs/{key}` | GET | Get config value by key |
| `/api/configs/{key}` | PUT | Set config value |
| `/api/configs/{key}` | DELETE | Delete config key |
| `/api/config-yaml` | GET | Get config YAML content |
| `/api/config-yaml` | PUT | Set config YAML content |

### 2.3 Data Refresh Strategy

Auto-polling via TanStack Query with context-aware intervals:

| Data | Refetch Interval | Condition |
|------|-------------------|-----------|
| Status (Claude/Cycle) | 5s / 15s | 5s when cycle running, 15s when idle |
| Tasks | 15s | Always |
| Messages | 10s | Always |
| Single Message | 5s | When viewing detail |
| Message Status | 5s | Always |
| Project Stats | 30s | Always |
| Health | 30s | Always |
| Auth Status | 30s | Always |

---

## 3. Authentication

### 3.1 Setup Page (`/setup`)

Multi-step initial setup wizard displayed on first access.

```
┌─────────────────────────────────┐
│  Claribot Setup                 │
│                                 │
│  Step [1] ─ [2] ─ [3]          │
│  (progress bar indicators)      │
│                                 │
│  ── Step 1: Set Password ────── │
│  Password:     [••••••••]       │
│  Confirm:      [••••••••]       │
│                         [Next]  │
│                                 │
│  ── Step 2: TOTP Setup ──────── │
│  Scan QR code with your         │
│  authenticator app:             │
│  ┌─────────┐                   │
│  │ [QR Code]│  (QRCodeSVG)     │
│  └─────────┘                   │
│  Google Authenticator or other  │
│  TOTP app                      │
│                         [Next]  │
│                                 │
│  ── Step 3: Verify TOTP ──────  │
│  Enter 6-digit code:            │
│  [______]  (numeric only)       │
│                      [Verify]   │
│  [QR code again]                │
└─────────────────────────────────┘
```

**Implementation:**
- Step 1: `POST /api/auth/setup` with `{ password }` → returns `{ totp_uri }`
- Step 2: Display QR code using `QRCodeSVG` from `qrcode.react`
- Step 3: `POST /api/auth/setup` with `{ password, totp_code }` → completes setup
- Password minimum 4 characters
- TOTP input: numeric-only, auto-strips non-digits, max 6 chars
- Step indicator: 3 progress bar segments

### 3.2 Login Page (`/login`)

```
┌─────────────────────────────────┐
│  Claribot Login                 │
│                                 │
│  Password:                      │
│  [••••••••]                     │
│                                 │
│  TOTP Code:                     │
│  [123456]  (centered, tracking) │
│                                 │
│  [Error message if failed]      │
│                                 │
│                        [Login]  │
└─────────────────────────────────┘
```

**Features:**
- Password + TOTP 6-digit code login via `POST /api/auth/login`
- Numeric-only TOTP input (filters non-digit characters, max 6)
- Enter key navigation: password field → focus TOTP, TOTP field → submit
- Error display on failed login
- TOTP input styled with `text-center text-lg tracking-widest`

### 3.3 Authentication Routing Guard

Implemented in `App.tsx` with `AuthGuard` component:

```
App Start
  │
  ├─ GET /api/auth/status
  │    ├─ Loading ──▶ Show spinner (Loader2 animate-spin)
  │    ├─ Error ──▶ Show "Cannot connect to server" message
  │    ├─ setup_completed = false ──▶ Redirect to /setup
  │    ├─ is_authenticated = false ──▶ Redirect to /login
  │    └─ is_authenticated = true ──▶ Render main app (Layout)
  │
  └─ Routes:
       /setup ──▶ Setup (no guard)
       /login ──▶ Login (no guard)
       /      ──▶ AuthGuard → Layout → child routes
       /*     ──▶ Redirect to /
```

- All protected routes are nested inside `<Layout>` via React Router outlet
- Logout button in Header triggers `POST /api/auth/logout` and invalidates auth query
- Auth hooks: `useAuthStatus`, `useLogin`, `useLogout`, `useSetup`, `useSetupVerify` (in `useAuth.ts`)

---

## 4. Page Layout

### 4.1 Overall Layout

```
Desktop:
┌──────────────────────────────────────────────────┐
│  Header: [≡]mobile Logo / ProjectSelector /      │
│          GlobalNav(desktop) / Claude Badge /      │
│          Connection Status / Logout               │
├──────────┬───────────────────────────────────────┤
│          │                                       │
│ Sidebar  │           Main Content                │
│ (220px)  │           (Outlet)                    │
│          │                                       │
│ [Project │                                       │
│  Card]   │                                       │
│ Edit     │                                       │
│ Specs    │                                       │
│ Tasks    │                                       │
│          │                                       │
└──────────┴───────────────────────────────────────┘

Mobile:
┌──────────────────────────────┐
│  Header: [≡] Logo [Badge]   │
├──────────────────────────────┤
│                              │
│        Main Content          │
│                              │
└──────────────────────────────┘
   ↓ Hamburger opens drawer
┌──────────┐
│ Sidebar  │
│ (overlay)│
│ Global:  │
│ Dashboard│
│ Messages │
│ Projects │
│ Schedules│
│ Settings │
│ ──────── │
│ Project: │
│ Specs    │
│ Tasks    │
└──────────┘
```

**Header Components (Header.tsx):**
- Left: Hamburger menu button (mobile only, min 44x44px) + Claribot logo (icon on mobile, text on sm+)
- Center-left: ProjectSelector dropdown (compact on mobile, hidden text on xs)
- Center: Global navigation links (desktop only, icons on md, icons+text on lg)
- Right: Claude status badge (`X/Y`) + connection status badge (hidden on mobile), logout button
- Navigation items: Dashboard, Messages, Projects, Schedules, Settings (global); Specs, Tasks (project-specific, in mobile drawer)

**Sidebar Components (Sidebar.tsx):**
- "Project" section header
- Current project card (when project selected, not GLOBAL):
  - Project name with folder/spinning icon (if cycle running)
  - Category badge
  - Status count badges (todo, planned, done, failed)
  - Stacked color bar (green/yellow/gray/red)
  - Progress percentage text
- Navigation: Edit (dynamic link to `/projects/{id}/edit`), Specs, Tasks
- Collapse/expand toggle button (desktop)
- Hidden on mobile (drawer mode via hamburger)

**ProjectSelector Component (ProjectSelector.tsx):**
- Dropdown trigger: folder icon + current project name (truncated, hidden on xs) + chevron
- Dropdown panel (320px wide, absolute positioned):
  - Search input with icon
  - Sort controls: cycle through last_accessed/created_at/task_count, toggle asc/desc
  - Category filter buttons (All + dynamic categories)
  - GLOBAL option at top
  - Project list with: pin toggle, project ID, category badge, description, inline category selector (on hover)
  - Outside click detection to close
  - ScrollArea with max-height 300px

---

### 4.2 Dashboard

**Path**: `/`

```
┌─────────────────────────────────────────────────────┐
│  Dashboard                                           │
├────────────┬────────────┬────────────┬──────────────┤
│ Claude     │ Cycle      │ Messages   │ Schedules    │
│ ● 2/10    │ ▶ Running   │ 3 process. │ 5 active     │
│ Running   │ PlanAll    │ 47 complet.│ 8 total      │
│           │ Task #12   │            │              │
│           │ 3m 24s     │            │              │
└────────────┴────────────┴────────────┴──────────────┘
│                                                      │
│  ── Recent Messages ──────────────────────────── [→] │
│  [done]  [cli]  Fix login bug                        │
│  [processing] [telegram] Review the code             │
│  [pending] [gui] Run tests                           │
│                                                      │
├──────────────────────────────────────────────────────┤
│  ── Projects ────────────────────────────────────── │
│                                                      │
│  ┌─────────────────┐  ┌─────────────────┐          │
│  │ ↻ claribot      │  │ blog             │          │
│  │ [backend]       │  │ Personal blog    │          │
│  │ 12 todo 5 plan  │  │ 3 todo 2 done    │          │
│  │ ████████░░ 75%  │  │ ██████░░░░ 50%  │          │
│  │ Done/Task: 80/106│  │ Done/Task: 4/8   │          │
│  │ [Edit][Tasks]   │  │ [Edit][Tasks]    │          │
│  │      [Stop]     │  │      [Cycle]     │          │
│  └─────────────────┘  └─────────────────┘          │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**Components:**
1. **4 Summary Cards** (responsive grid: 2cols on md, 4cols on lg):
   - Claude: Used/max count, Running/Idle status
   - Cycle: Status (idle/running/interrupted), type, phase, current task ID, elapsed time. Running icon animates.
   - Messages: Processing count, completed count
   - Schedules: Active count, total count
2. **Recent Messages**: Latest 5 messages with status badge (done/processing/pending), source label, truncated content. Arrow button to navigate to Messages page.
3. **Project Stats Board**: Per-project cards (responsive grid 1/2/3 cols) showing:
   - Project name with spinning icon if cycle running, category badge
   - Description (truncated)
   - Status count badges (todo, split, planned, done, failed)
   - Stacked status color bar (green/yellow/gray/red)
   - Progress bar with done/leaf ratio and percentage
   - Action buttons: Edit, Tasks, Cycle (or Stop if running)

**Data Refresh**: Auto-polling via TanStack Query (status: 5-15s, project stats: 30s)

---

### 4.3 Project Management (Projects)

**Path**: `/projects`

```
┌─────────────────────────────────────────────────────┐
│  Projects                           [+ Add Project]  │
│  [🔍 Search...]  [Sort: Last Accessed ▼] [↕]        │
│  [All | backend | frontend | ...]                    │
├─────────────────────────────────────────────────────┤
│                                                      │
│  ┌──────────────────────┐  ┌──────────────────────┐ │
│  │ 📌 claribot    [pin] │  │ blog           [pin] │ │
│  │ [backend]            │  │ Personal blog        │ │
│  │ 12 todo 5 planned    │  │ 3 todo 2 done        │ │
│  │ ████████░░ 75%       │  │ ██████░░░░ 50%       │ │
│  │ Done/Task: 80/106    │  │ Done/Task: 4/8       │ │
│  │ [Edit][Tasks][Cycle] │  │ [Edit][Tasks][Cycle]  │ │
│  └──────────────────────┘  └──────────────────────┘ │
│                                                      │
└─────────────────────────────────────────────────────┘
```

**Features:**
- Project card grid (responsive: 1/2/3 columns)
- Search by project ID, description, category
- Sort options: last_accessed, created_at, task_count (cycle button + asc/desc toggle)
- Category filter with dynamic category buttons
- Pin/unpin projects (pinned appear first, pin icon visible on hover)
- Per-project: status count badges, stacked color bar, progress bar with done/leaf ratio
- Action buttons: Edit (navigate to edit page), Tasks (switch + navigate), Cycle/Stop
- Running projects show spinning icon + Stop button instead of Cycle
- Add project form: Input accepts path (with `/`) for existing folder or ID for new project. Description textarea. Category selection with dynamic creation (+New).

**Project Edit Page** (`/projects/:id/edit`):
- Back button to projects list
- Read-only project ID and path display
- Editable: description (textarea), category (button group with dynamic creation), parallel Claude count (1-10 number input)
- Save button
- Danger Zone: Delete with confirmation (type project ID to confirm)

---

### 4.4 Task Management (Tasks) - Core Page

**Path**: `/tasks`

This is the core page of the Claribot web UI. It visually manages the Task tree structure with a 1:1 split-panel layout.

#### 4.4.1 Status Bar

```
┌────────────────────────────────────────────────────┐
│  Cycle: ▶ Running PlanAll Task #12 4/8  3m 24s    │
│  ●todo:12  ●split:20  ●planned:5  ●done:80        │
│  ●failed:2           done/leaf: 80/106 (75%)      │
└────────────────────────────────────────────────────┘
```

**Features:**
- **Cycle status row** (visible when not idle): Running/Interrupted indicator, type, phase badge, current task ID, completed/target count, elapsed time
- **Status counts row**: Clickable status filter buttons (colored dots with counts). Click to filter, click again to clear. `done/leaf` ratio with percentage on the right.
- Each status button: colored dot (gray/blue/yellow/green/red) + status name + count

#### 4.4.2 Task Tree View

```
┌──────────────────────────────────┬──────────────────────────────────┐
│  Tasks                           │  Task #4                     [×] │
│  [Tree|List]                     │  Handler Separation              │
│  [+] [Plan] [Run] [Cycle] [Stop]│                                  │
├──────────────────────────────────┤  Status: ● planned               │
│  [Status bar with filter]        │  depth: 2 | leaf                 │
│                                  │                                  │
│  ▼ #138 docs/webui.md update    │  [▶ Plan] [▶ Run] [Delete]       │
│  │              ● planned        │                                  │
│  │                               │  ── [Spec] [Plan] [Report] ───── │
│  ├─ ▼ #2 Router Modular.        │                                  │
│  │  │          ● split           │  ## Implementation Approach      │
│  │  ├── #4 Handler Sep.          │  1. Extract each command         │
│  │  │        ● planned           │     handler from router.go       │
│  │  └── #5 Add Middleware        │  2. Create handler/project.go    │
│  │           ● planned           │  3. Create handler/task.go       │
│  │                               │                                  │
│  ├── #3 Error Handling           │  [Edit]                          │
│  │        ● done                 │                                  │
│  └── #6 Write Tests              │                                  │
│           ● todo                 │                                  │
└──────────────────────────────────┴──────────────────────────────────┘
```

**Layout:** 1:1 split (w-1/2 each) on desktop, full-width on mobile.

**Toolbar:**
- View toggle: Tree / List (icon buttons)
- Add button (+)
- Bulk action buttons: Plan, Run, Cycle, Stop (text hidden on mobile, icons only)
- Action status indicator (yellow bar with spinner when pending)

**Tree Interactions:**
- Click row: Open Task detail panel (right panel)
- Expand/collapse nodes (chevron button)
- Tree indentation: `depth * 12 + 8px` padding
- Status dot + `#id` + title per row
- Selected row highlighted with `bg-accent`

**Sorting:** Tasks displayed newest-first (descending by ID).

**Status Icons:**
- `○` todo (gray-400)
- `◐` split (blue-400)
- `●` planned (yellow-400)
- `▶` running (blue, animated) — conceptual, shown via cycle status
- `✅` done (green-400)
- `❌` failed (red-400)

#### 4.4.3 List View

```
┌──────────────────────────────────────────────────────────┐
│  Desktop: Table view                                      │
├────┬───────────────────┬────────────┬─────┬──────────────┤
│ ID │ Title             │ Status     │Depth│ Parent       │
├────┼───────────────────┼────────────┼─────┼──────────────┤
│#138│ docs/webui.md up..│ ● planned  │ 0   │ -            │
│ #7 │ Doc Update        │ ● todo     │ 0   │ -            │
│ #6 │ Write Tests       │ ● todo     │ 1   │ #1           │
└────┴───────────────────┴────────────┴─────┴──────────────┘

  Mobile: Card view
┌──────────────────────────────────┐
│  #138                  ● planned │
│  docs/webui.md update            │
│  depth: 0    parent: -           │
└──────────────────────────────────┘
```

**Desktop:** Table with columns: ID, Title, Status (dot), Depth, Parent.
**Mobile:** Card view with ID + status dot header, title, depth + parent info.
**Sorting:** Newest-first (descending by ID).

#### 4.4.4 Task Detail Panel (Right Panel, 1:1 Split)

```
┌───────────────────────────────┐
│  Task #4                  [×] │
├───────────────────────────────┤
│                               │
│  Handler Separation           │
│  ● planned                    │
│  depth: 2 | leaf              │
│                               │
│  [▶ Plan] [▶ Run] [Delete]   │
│                               │
│  ── [Spec] [Plan] [Report] ── │
│  (Tab-based navigation)       │
│                               │
│  Content rendered as HTML     │
│  from markdown source.        │
│                               │
│  [Edit] (toggle textarea)    │
│                               │
└───────────────────────────────┘
```

**Features:**
- Title and status dot with depth info and leaf badge
- Action buttons: Plan, Run, Delete (with confirm dialog)
- Tab-based Spec/Plan/Report switching (3-column grid TabsList)
- Markdown → HTML rendering via MarkdownRenderer component
- Inline editing: Click Edit → textarea with Save/Cancel buttons
- ScrollArea for content overflow

**Mobile:** Detail panel opens as a full-screen overlay (`fixed inset-0 z-50`) instead of side panel.

#### 4.4.5 Task Action Buttons

| Button | Action | API Call |
|--------|--------|---------|
| `+ Add Task` | Create new Task form (title, parent_id, spec) | `POST /api/tasks` |
| `Plan` (bulk) | Generate plans for all todo Tasks | `POST /api/tasks/plan-all` |
| `Run` (bulk) | Execute all planned leaf Tasks | `POST /api/tasks/run-all` |
| `Cycle` (bulk) | Auto-cycle Plan + Run | `POST /api/tasks/cycle` |
| `Stop` (bulk) | Stop active cycle | `POST /api/tasks/stop` |
| `▶ Plan` (individual) | Generate plan for single Task | `POST /api/tasks/{id}/plan` |
| `▶ Run` (individual) | Execute single Task | `POST /api/tasks/{id}/run` |

---

### 4.5 Messages (Chat UI)

**Path**: `/messages`

Redesigned as a chat interface with 1:1 split-panel layout.

```
┌─────────────────────┬───────────────────────────────────┐
│  Messages           │  Message #47 Detail                │
│                     │                                    │
├─────────────────────┤  [←](mobile) Message #47           │
│                     │  [done] [📨 telegram]  14:23       │
│  ── Today ────────  │                                    │
│                     │  ── Content ──────────────────────  │
│      Organize the   │  Organize the API endpoint list     │
│      API endpoint ⌐ │                                    │
│                     │  ── Result ───────────────────────  │
│  ⌐ Completed        │  ## API Endpoint List               │
│    [done] ✅        │  1. POST /api - Execute command     │
│    API List 1.POST..│  2. GET /api/health - Health check   │
│    [Detail]         │  ...                                │
│              14:23  │                                    │
│                     │  ── Error (if any) ────────────── │
│      Review the code│  (red background, pre-formatted)   │
│                     │                                    │
│  ⌐ Processing...    │                                    │
│    [processing] 🔄  │                                    │
│              14:20  │                                    │
│                     │                                    │
│  ── Yesterday ────  │                                    │
│                     │                                    │
│      Run the tests  │                                    │
│  ⌐ All tests pass   │                                    │
│              09:00  │                                    │
│                     │                                    │
├─────────────────────┤                                    │
│ [Message input...   │                                    │
│       (Ctrl+Enter)] │                                    │
│              [Send] │                                    │
└─────────────────────┴───────────────────────────────────┘
```

**Features:**
- **Chat bubble UI** (ChatBubble component): User messages (right-aligned, primary bg, rounded-br-md) and bot responses (left-aligned, muted bg, rounded-bl-md) in conversation pairs
- **Date grouping**: Messages grouped by date headers with horizontal line separators
- **Source labels**: Telegram, CLI, Schedule shown above user bubbles
- **Bot bubbles**: Status indicator badge (pending/processing/done/failed), result summary (first 1-2 lines), "Detail" link
- **Message input**: Textarea at bottom with Send button, Ctrl+Enter/Cmd+Enter shortcut, 2 rows
- **Optimistic updates**: Pending messages shown immediately with temporary ID before server confirmation, removed on success/error
- **Detail panel**: Right panel (1:1 split) shows full message content + result with MarkdownRenderer, error in red pre block
- **Auto-scroll**: Chat scrolls to latest message on entry (instant, no animation), smooth on subsequent messages
- **Message sorting**: Ascending by created_at (oldest first for chat flow)

**Mobile:** Single-pane view, switches between chat list and detail view via state toggle. Back button on detail view.

---

### 4.6 Specs

**Path**: `/specs`

```
┌──────────────────────┬──────────────────────────────────┐
│  Specs         [+]   │  Spec #3                     [×] │
│  [🔍 Search...]      │  User Authentication Flow        │
│  [all|draft|review|  │  [Edit Title] [Delete]           │
│   approved|deprecated│                                  │
├──────────────────────┤  [Approved] priority: 4           │
│                      │  2025-01-12                       │
│  ┌────────────────┐  │                                  │
│  │ #5 API Redesign│  │  ── Status ─────────────────── │
│  │  [Draft] P:3   │  │  [draft][review][approved]      │
│  │  API redesign..│  │  [deprecated]                   │
│  └────────────────┘  │                                  │
│                      │  ── Priority ─────────────────  │
│  ┌────────────────┐  │  [1][2][3][4][5]                │
│  │ #3 Auth Flow   │  │                                  │
│  │  [Approved] P:4│  │  ── Content ──────────────────  │
│  │  Requirements..│  │  [Preview | Edit]               │
│  └────────────────┘  │                                  │
│                      │  ## Requirements                 │
│  ┌────────────────┐  │  1. Password + TOTP login        │
│  │ #1 Init Setup  │  │  2. Session management           │
│  │  [Deprecated]  │  │                                  │
│  │  P:1           │  │                                  │
│  └────────────────┘  │                                  │
└──────────────────────┴──────────────────────────────────┘
```

**Features:**
- **Dual-panel layout**: List (1/3) + Detail (2/3) on desktop
- **List panel**: Card-based layout (not table) with ID, status badge, title, priority, date, content preview (2-line clamp)
- **Add spec form**: Title input + content markdown textarea
- **Search**: Filter specs by title and content keyword
- **Status filter**: Buttons for all, draft, review, approved, deprecated (with counts)
- **Detail panel**:
  - Title with inline edit (Edit/Delete buttons)
  - Status badge + priority + date
  - Status change: Button group for all 4 statuses
  - Priority change: Button group [1-5]
  - Content: Preview/Edit toggle. Preview shows MarkdownRenderer, Edit shows auto-sizing textarea with Save/Cancel
- **Delete**: With `confirm()` dialog
- Spec items sorted ascending by ID

**Status Badges:**
- `Draft` (secondary/gray)
- `Review` (warning/yellow)
- `Approved` (success/green)
- `Deprecated` (destructive/red)

**Mobile:** Full-screen overlay for detail view (`fixed inset-0 z-50`).

---

### 4.7 Schedules

**Path**: `/schedules`

```
┌─────────────────────────────────────────────────────┐
│  Schedules                          [+ Add Schedule] │
├─────────────────────────────────────────────────────┤
│                                                      │
│  ┌───────────────────────────────────────────────┐  │
│  │ #1  [ON]  [🤖 Claude]  [run_once]  [blog]    │  │
│  │ ⏰ 0 9 * * 1-5  (Weekdays 09:00)             │  │
│  │ "Review code changes and organize issues"     │  │
│  │ Last: 01-15 09:00  Next: 01-16 09:00         │  │
│  │                          [⏻][📋][🗑]         │  │
│  │  ── Run History ──────────────────────────    │  │
│  │  #15 🤖 [done] 01-15 09:00                   │  │
│  │  Result: ## Code Review...                    │  │
│  └───────────────────────────────────────────────┘  │
│                                                      │
│  ┌───────────────────────────────────────────────┐  │
│  │ #2  [ON]  [💻 Bash]                           │  │
│  │ ⏰ */5 * * * *  (Every 5h)                    │  │
│  │ "curl -s https://api.example.com/health"      │  │
│  └───────────────────────────────────────────────┘  │
│                                                      │
└─────────────────────────────────────────────────────┘
```

**Add Schedule Form (Card):**
- Cron expression input with human-readable preview (min hour day month weekday hint)
- Type selector: `<select>` with Claude (AI) / Bash (Command) options
- Message/command textarea (placeholder changes based on type)
- Project selector dropdown (Global + all projects)
- Run once checkbox

**Features:**
- Schedule card list with badges: ON/OFF, type (🤖 Claude / 💻 Bash), run_once, project
- Vertical action buttons (right side): Toggle enable/disable, History, Delete
- Run history viewer (expandable per schedule via History button)
- Run history shows: run ID, type icon, status badge, timestamp, result/error in pre block
- Cron expression with `<code>` styling + human-readable description
- Last run / Next run timestamps
- Mobile: Action buttons vertically stacked, cron expression in scrollable container

---

### 4.8 Settings

**Path**: `/settings`

```
┌─────────────────────────────────────────────────┐
│  Settings                                        │
├─────────────────────────────────────────────────┤
│                                                  │
│  ── 🖥 System Info ────────────────────────────  │
│  Claribot Version: v0.2.32                       │
│  Uptime: 3d 14h 22m                             │
│  DB Path: ~/.claribot/db.clt                    │
│  Service: ● Connected                            │
│                                                  │
│  ── 📄 Config (config.yaml) ──────────────────  │
│  Edit ~/.claribot/config.yaml. Restart to apply. │
│                                                  │
│  ⚙ Service                                      │
│    Host: [127.0.0.1]    (default: 127.0.0.1)    │
│    Port: [9847]         (default: 9847)          │
│                                                  │
│  💬 Telegram                                     │
│    Bot Token: [•••••••••••]                      │
│    Admin Chat ID: [123456789]                    │
│    Allowed Users: [123, 456, 789]                │
│                                                  │
│  🤖 Claude                                       │
│    Max Concurrent: [10]  (1-10, default: 10)     │
│    Idle Timeout: [1200]  (default: 1200 = 20min) │
│    Max Timeout: [1800]   (60-7200, default: 1800)│
│                                                  │
│  📁 Project                                      │
│    Default Path: [/home/user/projects]           │
│                                                  │
│  📋 Pagination                                   │
│    Page Size: [10]       (1-100, default: 10)    │
│                                                  │
│  📜 Log                                          │
│    Level: [info ▼]  (debug/info/warn/error)      │
│    File: [~/.claribot/claribot.log]              │
│                                                  │
│                    [Save Config]                  │
│                                                  │
└─────────────────────────────────────────────────┘
```

**Features:**
- System info display: version (from `/api/health`), uptime (formatted days/hours/min), DB path, connection status badge
- Config YAML editor organized by section with icons:
  - Service (host, port)
  - Telegram (token as password, admin_chat_id, allowed_users as comma-separated)
  - Claude (max concurrent, idle timeout, max timeout)
  - Project (default path)
  - Pagination (page size)
  - Log (level dropdown, file path)
- Live config save/load via API (`GET/PUT /api/config-yaml`)
- YAML parsing with `yaml` npm package
- Smart save: only includes non-default values in saved YAML
- Success/error feedback messages after save

---

## 5. Mobile Responsiveness

Comprehensive mobile optimization has been implemented across all pages and components.

### 5.1 Breakpoints
| Breakpoint | Width | Usage |
|------------|-------|-------|
| `sm` | 640px | Small adjustments |
| `md` | 768px | Tablet layout switch |
| `lg` | 1024px | Full desktop layout |

### 5.2 Mobile Optimizations by Component

| Component | Mobile Behavior |
|-----------|-----------------|
| **Header** | Hamburger menu (min 44x44px) replaces nav links, Claude/connection badges hidden, ProjectSelector shows icon only |
| **Sidebar** | Hidden by default, opens as Sheet overlay drawer via hamburger, nav items with 44px touch targets |
| **Layout** | Reduced padding (`p-3 sm:p-4 md:p-6`) |
| **Dashboard** | Grid collapses to 2 columns (md) then 1 column |
| **Tasks** | Detail panel opens as full-screen overlay (`fixed inset-0 z-50`); tree indentation reduced (`depth*12`); card view instead of table on mobile; toolbar text hidden (icons only); Plan/Run/Cycle/Stop buttons use `flex-1 min-w-0` |
| **Messages** | Single-pane mode with state toggle between chat and detail view, Back button on detail |
| **Specs** | Full-screen overlay for detail view, card-based list |
| **Schedules** | Action buttons vertically stacked, cron expression scrollable |
| **Projects** | Grid collapses to single column |
| **Page titles** | Responsive font size (`text-2xl md:text-3xl`) |

### 5.3 Touch Targets
- All interactive elements maintain minimum 44x44px touch targets
- Buttons: `min-h-[44px]` and `min-w-[44px]` on mobile
- Input fields, badges, and navigation links sized for touch
- Sidebar navigation items: `py-3` for adequate touch height

---

## 6. Common UI Components

### 6.1 MarkdownRenderer

Uses `react-markdown` with `remark-gfm` plugin for rendering markdown content as HTML. Wraps output in `markdown-body` class div.

Used in:
- Task Spec/Plan/Report tabs
- Message results
- Spec content preview

### 6.2 ChatBubble

Reusable chat message bubble component with props: type (user/bot), content, status, source, result, time, onDetailClick, isSelected.

- User bubbles: Right-aligned, primary background, rounded corners (rounded-br-md for tail)
- Bot bubbles: Left-aligned, muted background, status indicator badge, result summary (2-line clamp), detail link
- Source label above user bubbles (Telegram/CLI/Schedule)
- Timestamp below bubbles

### 6.3 Confirmation Dialog

Displayed for dangerous operations (delete Task, delete Project, delete Schedule). Uses browser `confirm()` for simple operations, custom UI with type-to-confirm for project deletion.

### 6.4 Status Badges

Consistent color-coded status indicators across all pages:

| Status | Color | Usage |
|--------|-------|-------|
| todo | Gray | Task |
| planned | Yellow | Task |
| split | Blue | Task |
| done | Green | Task, Message, Schedule Run |
| failed | Red | Task, Message, Schedule Run |
| pending | Gray | Message |
| processing | Yellow | Message |
| draft | Gray/Secondary | Spec |
| review | Yellow/Warning | Spec |
| approved | Green/Success | Spec |
| deprecated | Red/Destructive | Spec |

---

## 7. Deployment & Integration

### 7.1 Go embed Integration

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

### 7.2 HTTP Routing

```go
// RESTful router (bot/internal/handler/restful.go)
// Auth endpoints (no middleware)
POST /api/auth/setup
GET  /api/auth/totp-setup
POST /api/auth/login
POST /api/auth/logout
GET  /api/auth/status

// Protected endpoints (auth middleware)
GET  /api/health
GET  /api/status
GET  /api/usage
POST /api/usage/refresh
GET/PUT /api/config-yaml
GET/PUT/DELETE /api/configs/{key}
// ... all resource endpoints (projects, tasks, messages, schedules, specs)

// Static file serving (SPA fallback)
GET  /*  → webui.Handler() → index.html
```

**SPA Fallback**: All requests not starting with `/api` are redirected to `index.html` (React Router support).

### 7.3 Makefile Additions

```makefile
build-gui:
	cd gui && npm install && npm run build
	rm -rf bot/internal/webui/dist
	cp -r gui/dist bot/internal/webui/dist

build: build-gui build-cli build-bot

dev-gui:
	cd gui && npm run dev
```

---

## 8. Implementation Status

### Phase 1: Foundation (MVP) ✅
1. ~~Project scaffolding (Vite + React + TypeScript + shadcn/ui)~~ ✅
2. ~~API client module (RESTful API communication)~~ ✅
3. ~~Layout component (Header + Sidebar + Main)~~ ✅
4. ~~ProjectSelector (dropdown with search, sort, category filter, pin)~~ ✅
5. ~~Dashboard page (summary cards + project stats board)~~ ✅
6. ~~Go embed integration and static file serving~~ ✅

### Phase 2: Core Features ✅
7. ~~Projects page (CRUD + search + sort + category + pin)~~ ✅
8. ~~ProjectEdit page (edit description, parallel, category, delete)~~ ✅
9. ~~Tasks page - List view (status filter, priority display, newest-first sort)~~ ✅
10. ~~Tasks page - Tree view (expand/collapse, status dots)~~ ✅
11. ~~Task detail panel (Spec/Plan/Report tabs, markdown rendering, inline editing)~~ ✅
12. ~~Task execution buttons (Plan/Run/Cycle/Stop)~~ ✅
13. ~~Task status bar (status counts + cycle progress)~~ ✅
14. ~~Messages page (chat UI with bubbles, date groups, optimistic updates)~~ ✅

### Phase 3: Visualization ✅
15. ~~Dashboard project stats board (per-project progress bars)~~ ✅
16. ~~Cycle status display (phase, progress, target count)~~ ✅

### Phase 4: Authentication ✅
17. ~~Setup page (multi-step: password → TOTP QR → verify)~~ ✅
18. ~~Login page (password + TOTP 6-digit code)~~ ✅
19. ~~Auth routing guard (App.tsx: setup check → login check → render)~~ ✅
20. ~~Logout functionality (Header button)~~ ✅

### Phase 5: Extended Features ✅
21. ~~Schedules page (CRUD + type selector: Claude/Bash + run history)~~ ✅
22. ~~Settings page (system info + config YAML editor)~~ ✅
23. ~~Specs page (CRUD + search + status filter + priority + markdown editor)~~ ✅

### Phase 6: Mobile Responsiveness ✅
24. ~~Sidebar hamburger drawer~~ ✅
25. ~~Header responsive (badge hiding, ProjectSelector compact)~~ ✅
26. ~~Layout padding responsive~~ ✅
27. ~~Tasks detail panel mobile overlay~~ ✅
28. ~~Tasks table → card view on mobile~~ ✅
29. ~~Tasks tree indentation reduced~~ ✅
30. ~~Tasks toolbar button wrap~~ ✅
31. ~~Schedules card responsive~~ ✅
32. ~~Page title font responsive~~ ✅
33. ~~Touch targets minimum 44x44px~~ ✅

### Phase 7: Not Yet Implemented
34. WebSocket integration (`/api/stream`) for real-time updates
35. Real-time Claude execution log streaming
36. Dark mode toggle (CSS variables ready)
37. Keyboard shortcuts (Task navigation, execution)

---

## 9. Screen Flow Diagram

```
[App Start]
  │
  ├─ needs_setup ──▶ [Setup] ── complete ──▶ [Login]
  │
  ├─ not authenticated ──▶ [Login] ── success ──▶ [Dashboard]
  │
  └─ authenticated ──▶ [Dashboard]
                            │
                            ├──▶ [Projects] ── Click Edit ──▶ [ProjectEdit]
                            │         └── Click Tasks ──▶ switch + [Tasks]
                            │
                            ├──▶ [Tasks]
                            │      ├── Tree/List View ── Click Task ──▶ [Task Detail Panel]
                            │      ├── Status Filter ── Click status dot
                            │      └── Plan/Run/Cycle/Stop Button ──▶ Trigger API
                            │
                            ├──▶ [Messages]
                            │      ├── Send Message ──▶ Optimistic + API
                            │      └── Click Detail ──▶ [Message Detail]
                            │
                            ├──▶ [Specs]
                            │      ├── Add/Edit Spec ──▶ Markdown Editor
                            │      ├── Status/Search Filter
                            │      └── Click Spec ──▶ [Spec Detail]
                            │
                            ├──▶ [Schedules]
                            │      ├── Add Schedule ──▶ [Add Form]
                            │      ├── ON/OFF Toggle
                            │      └── View History ──▶ [Run History]
                            │
                            └──▶ [Settings]
                                   └── Edit Config ──▶ Save YAML

[Header ProjectSelector] ── Select Project ──▶ Switch + Invalidate All Queries
[Header Logout] ──▶ Invalidate auth ──▶ [Login]
[Sidebar Edit] ──▶ [ProjectEdit]
```

---

## 10. Type Definitions

Key TypeScript types used across the GUI (`gui/src/types/index.ts`):

| Type | Fields | Usage |
|------|--------|-------|
| `Project` | id, name, path, type, description, status, category, pinned, last_accessed, created_at, updated_at | Projects page, ProjectSelector |
| `Task` | id, parent_id, title, spec, plan, report, status, error, is_leaf, depth, created_at, updated_at | Tasks page |
| `Message` | id, project_id, content, source, status, result, error, created_at, completed_at | Messages page |
| `Schedule` | id, project_id, cron_expr, message, type, enabled, run_once, last_run, next_run, created_at, updated_at | Schedules page |
| `ScheduleRun` | id, schedule_id, status, result, error, started_at, completed_at | Schedule history |
| `Spec` | id, title, content, status, priority, created_at, updated_at | Specs page |
| `ClaudeStatus` | used, max, available | Status polling |
| `CycleStatus` | status, type, project_id, started_at, current_task_id, active_workers, phase, target_total, completed, elapsed_sec | Dashboard, Tasks |
| `TaskStats` | total, leaf, todo, planned, split, done, failed | Dashboard, Sidebar |
| `ProjectStats` | project_id, project_name, project_description, stats (TaskStats & { in_progress }) | Dashboard |
| `StatusResponse` | success, message, data (ClaudeStatus), cycle_status, cycle_statuses[], task_stats | Status polling |
| `PaginatedList<T>` | items, total, page, page_size, total_pages | List API responses |
| `UsageData` | success, message, live?, updated_at? | Usage API (client.ts) |
