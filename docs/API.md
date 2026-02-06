# Claribot API Documentation

## Overview

Claribot provides a RESTful HTTP API for CLI, Web UI, and Telegram integration. The default port is `9847`.

- **Base URL**: `http://127.0.0.1:9847`
- **Content-Type**: `application/json`
- **Middleware**: CORS → Auth → Router

## Authentication

Remote (non-localhost) requests require authentication via JWT cookie. Localhost requests (`127.0.0.1`, `::1`) bypass authentication.

### Auth Endpoints

Auth endpoints are exempt from the auth middleware.

#### POST /api/auth/setup

Initial password setup. Called once when the system is first configured.

**Request**:
```json
{
  "password": "your-password",
  "totp_code": "123456"
}
```

**Response** (first call - returns TOTP URI for QR setup):
```json
{
  "totp_uri": "otpauth://totp/Claribot?secret=..."
}
```

**Response** (second call with TOTP code - completes setup):
```json
{
  "success": true
}
```
Sets `auth_token` cookie on success.

#### POST /api/auth/login

Authenticate with password and TOTP code.

**Request**:
```json
{
  "password": "your-password",
  "totp_code": "123456"
}
```

**Response**:
```json
{
  "success": true
}
```
Sets `auth_token` cookie on success.

#### GET /api/auth/status

Check authentication status.

**Response**:
```json
{
  "setup_completed": true,
  "authenticated": true
}
```

#### POST /api/auth/logout

Clear authentication cookie.

**Response**:
```json
{
  "success": true
}
```

#### GET /api/auth/totp-setup

Get TOTP setup URI for QR code generation.

**Response**:
```json
{
  "totp_uri": "otpauth://totp/Claribot?secret=..."
}
```

## Response Format

All RESTful endpoints return this structure:

```json
{
  "success": true,
  "message": "Response message",
  "data": { ... },
  "needs_input": false,
  "prompt": "",
  "context": "",
  "error_type": ""
}
```

| Field | Type | Description |
|-------|------|-------------|
| success | bool | Whether the request succeeded |
| message | string | Message to display to the user |
| data | any | Structured response data (optional) |
| needs_input | bool | Whether additional input is required |
| prompt | string | Input prompt (when needs_input=true) |
| context | string | Context to include in the next request |
| error_type | string | Error type identifier (e.g., `auth_error`) (optional) |

## Pagination

All list endpoints support pagination via query parameters:

```
GET /api/projects?page=2&page_size=20
GET /api/tasks?all=true
```

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| page | int | 1 | Page number (starts from 1) |
| page_size | int | 10 | Items per page |
| all | bool | false | Fetch all items (ignores pagination) |

**Response wrapper**:
```json
{
  "items": [...],
  "page": 1,
  "page_size": 10,
  "total_items": 25,
  "total_pages": 3,
  "has_next": true,
  "has_prev": false
}
```

---

## Status & Usage

### GET /api/status

Get system status including Claude sessions, cycle status, and task statistics.

**Response**:
```json
{
  "success": true,
  "message": "...",
  "data": {
    "max": 3,
    "used": 1,
    "available": 2,
    "sessions": 1
  },
  "cycle_status": {
    "status": "idle",
    "type": "cycle",
    "project_id": "myproject",
    "started_at": "2025-01-01T00:00:00Z",
    "elapsed_sec": 120,
    "current_task_id": 5,
    "active_workers": 2,
    "phase": "running",
    "target_total": 10,
    "completed": 3
  },
  "cycle_statuses": [...],
  "task_stats": {
    "total": 50,
    "leaf": 25,
    "todo": 10,
    "planned": 5,
    "done": 30,
    "failed": 1,
    "in_progress": 2
  }
}
```

### GET /api/usage

Get Claude API usage statistics.

**Response**:
```json
{
  "success": true,
  "message": "formatted stats output",
  "live": "live usage output",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### POST /api/usage/refresh

Trigger async refresh of live usage data.

**Response**: `202 Accepted`
```json
{
  "success": true,
  "message": "Usage refresh started"
}
```

---

## Projects

### GET /api/projects

List all projects (paginated).

**Query Parameters**: `page`, `page_size`, `all`

**Response Data**:
```json
{
  "items": [
    {
      "id": "myproject",
      "path": "/path/to/project",
      "description": "My Project"
    }
  ],
  "page": 1,
  "total_items": 5,
  "total_pages": 1
}
```

### POST /api/projects

Create a new project.

**Request**:
```json
{
  "id": "project-id",
  "path": "/path/to/project",
  "description": "Project description"
}
```

### GET /api/projects/stats

Get task statistics for all projects.

**Response Data**:
```json
[
  {
    "project_id": "myproject",
    "project_name": "My Project",
    "project_description": "Project description",
    "stats": {
      "total": 20,
      "leaf": 15,
      "todo": 3,
      "planned": 1,
      "done": 10,
      "failed": 0,
      "in_progress": 1
    }
  }
]
```

### GET /api/projects/{id}

Get specific project details.

### PATCH /api/projects/{id}

Update project fields.

**Request** (field/value format):
```json
{
  "field": "parallel",
  "value": "2"
}
```

**Request** (direct format):
```json
{
  "description": "New description",
  "parallel": 2
}
```

**Supported Fields** (single field format):
- `description`: Project description
- `parallel`: Number of concurrent Claude instances
- `category`: Project category
- `pinned`: Pin/unpin project

### DELETE /api/projects/{id}

Delete a project.

### POST /api/projects/{id}/switch

Switch to a project (set as active context). All subsequent project-scoped API calls use this context.

---

## Tasks

Task endpoints require a project to be selected.

### GET /api/tasks

List tasks (supports tree view and flat list with pagination).

**Query Parameters**:

| Parameter | Type | Description |
|-----------|------|-------------|
| tree | bool | Return full task tree structure |
| parent_id | int | Filter by parent task ID |
| page | int | Page number |
| page_size | int | Items per page |
| all | bool | Fetch all items |

### POST /api/tasks

Create a new task.

**Request**:
```json
{
  "title": "Task title",
  "parent_id": 1,
  "spec": "Task specification"
}
```

**Response**: `201 Created`

### GET /api/tasks/{id}

Get specific task details (includes spec, plan, report).

### PATCH /api/tasks/{id}

Update task fields.

**Request**:
```json
{
  "field": "status",
  "value": "done"
}
```

**Updatable Fields**:
- `title`: Title
- `spec`: Specification
- `plan`: Execution plan
- `report`: Execution report
- `status`: Status (`todo`, `planned`, `split`, `done`, `failed`)
- `priority`: Execution order priority (integer)

### DELETE /api/tasks/{id}

Delete a task.

### POST /api/tasks/{id}/plan

Generate a plan for a single task using Claude.

### POST /api/tasks/{id}/run

Execute a single task using Claude.

### POST /api/tasks/plan-all

Generate plans for all eligible tasks.

### POST /api/tasks/run-all

Execute all eligible tasks.

### POST /api/tasks/cycle

Run full cycle (plan + run) for all tasks.

### POST /api/tasks/stop

Stop currently running task traversal.

---

## Messages

### GET /api/messages

List all messages (paginated).

**Query Parameters**: `page`, `page_size`, `all`

### POST /api/messages

Send a message to Claude.

**Request**:
```json
{
  "content": "Message content",
  "source": "gui"
}
```

**Source values**:
- `telegram`: Sent from Telegram
- `cli`: Sent from CLI
- `gui`: Sent from Web UI
- `api`: Default (when source is omitted)

### GET /api/messages/{id}

Get specific message details.

### GET /api/messages/status

Get message queue status.

**Response Data**:
```json
{
  "total": 100,
  "pending": 5,
  "processing": 2,
  "done": 90,
  "failed": 3
}
```

### GET /api/messages/processing

Get the currently processing message.

---

## Configs

Global configuration key-value store.

### GET /api/configs

List all config key-value pairs.

### GET /api/configs/{key}

Get specific config value.

### PUT /api/configs/{key}

Set a config value.

**Request**:
```json
{
  "value": "config-value"
}
```

### DELETE /api/configs/{key}

Delete a config entry.

### GET /api/config-yaml

Get raw config YAML file content.

### PUT /api/config-yaml

Update raw config YAML file.

**Request**:
```json
{
  "content": "key1: value1\nkey2: value2"
}
```

---

## Schedules

### GET /api/schedules

List schedules (filtered by current project context).

**Query Parameters**:

| Parameter | Type | Description |
|-----------|------|-------------|
| all | bool | Show schedules from all projects |
| project_id | string | Filter by project ID |

### POST /api/schedules

Create a new schedule.

**Request**:
```json
{
  "cron_expr": "0 9 * * *",
  "message": "Daily task",
  "type": "claude",
  "project_id": "myproject",
  "run_once": false
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| cron_expr | string | required | Cron expression |
| message | string | required | Command/message to execute |
| type | string | `"claude"` | Execution type: `claude` (Claude Code) or `bash` (shell command) |
| project_id | string | current | Target project (defaults to active project) |
| run_once | bool | false | Auto-disable after single execution |

**Response**: `201 Created`

### GET /api/schedules/{id}

Get specific schedule details.

### PATCH /api/schedules/{id}

Update schedule fields.

**Request**:
```json
{
  "field": "project",
  "value": "new-project-id"
}
```

**Updatable Fields**:
- `project`: Assign schedule to a project (use empty string or `"none"` to unassign)

### DELETE /api/schedules/{id}

Delete a schedule.

### POST /api/schedules/{id}/enable

Enable a schedule.

### POST /api/schedules/{id}/disable

Disable a schedule.

### GET /api/schedules/{id}/runs

List execution history for a schedule (paginated).

### GET /api/schedule-runs/{runId}

Get a specific execution record.

---

## Specs

Specification management for project requirements. Requires a project to be selected.

### GET /api/specs

List all specifications (paginated).

**Query Parameters**: `page`, `page_size`, `all`

### POST /api/specs

Create a new specification.

**Request**:
```json
{
  "title": "Specification title",
  "content": "Specification content"
}
```

**Response**: `201 Created`

### GET /api/specs/{id}

Get specific specification details.

### PATCH /api/specs/{id}

Update specification fields.

**Request**:
```json
{
  "field": "content",
  "value": "Updated content"
}
```

**Updatable Fields**:
- `title`: Specification title
- `content`: Specification content
- `status`: Status (`draft`, `review`, `approved`, `deprecated`)
- `priority`: Priority (integer)

### DELETE /api/specs/{id}

Delete a specification.

---

## Legacy API

For backward compatibility, the old command-style API is still available:

```
GET /?args=project+list
POST /api { "command": "project", "subcommand": "list" }
GET /api/health
```

These endpoints are used by the CLI (`clari`) and Telegram handler.

---

## Error Handling

| HTTP Status | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Resource created |
| 202 | Async operation accepted |
| 400 | Bad request (success=false) |
| 401 | Authentication required |
| 405 | Method not allowed |
| 500 | Internal server error |

## HTTP Timeout Settings

- **ReadTimeout**: 10 seconds
- **WriteTimeout**: 30 minutes (considering Claude execution time)
- **IdleTimeout**: 60 seconds
