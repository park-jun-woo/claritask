# Claribot API Documentation

## Overview

Claribot communicates with the CLI via HTTP API. The default port is `9847`.

- **Base URL**: `http://127.0.0.1:9847`
- **Content-Type**: `application/json`

## Request Format

### POST (Recommended)

```json
{
  "command": "project",
  "subcommand": "list",
  "args": ["arg1", "arg2"],
  "flags": {
    "page": 1,
    "pageSize": 10
  }
}
```

### GET (Backward Compatible)

```
GET /?args=project+list
```

## Response Format

```json
{
  "success": true,
  "message": "Response message",
  "data": { ... },
  "needs_input": false,
  "prompt": "",
  "context": ""
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

## Commands

### Project

Project management commands

#### project list

```
GET /?args=project+list
POST { "command": "project", "subcommand": "list" }
```

**Response Data**:
```json
{
  "items": [
    {
      "id": "myproject",
      "path": "/path/to/project",
      "type": "go",
      "description": "My Project"
    }
  ],
  "page": 1,
  "total_items": 5,
  "total_pages": 1
}
```

#### project add

```
POST { "command": "project", "subcommand": "add", "args": ["/path/to/project", "go", "Description"] }
```

#### project create

```
POST { "command": "project", "subcommand": "create", "args": ["project-id", "go", "Description"] }
```

#### project switch

```
POST { "command": "project", "subcommand": "switch", "args": ["project-id"] }
```

#### project get

```
POST { "command": "project", "subcommand": "get", "args": ["project-id"] }
```

#### project delete

```
POST { "command": "project", "subcommand": "delete", "args": ["project-id", "yes"] }
```

### Task

Task management commands. Requires a project to be selected.

#### task list

```
POST { "command": "task", "subcommand": "list", "flags": { "page": 1 } }
```

**Query Parameters**:
- `parent_id` (int, optional): Parent Task ID
- `-p` (int): Page number
- `-n` (int): Page size

#### task add

```
POST { "command": "task", "subcommand": "add", "args": ["Task Title", "--parent", "1"] }
```

#### task get

```
POST { "command": "task", "subcommand": "get", "args": ["1"] }
```

#### task set

```
POST { "command": "task", "subcommand": "set", "args": ["1", "status", "done"] }
```

**Fields**:
- `title`: Title
- `status`: Status (todo, in_progress, done)
- `description`: Description

#### task delete

```
POST { "command": "task", "subcommand": "delete", "args": ["1", "yes"] }
```

#### task plan

Generate a Task plan using Claude

```
POST { "command": "task", "subcommand": "plan", "args": ["1"] }
POST { "command": "task", "subcommand": "plan", "args": ["--all"] }
```

#### task run

Execute a Task using Claude

```
POST { "command": "task", "subcommand": "run", "args": ["1"] }
POST { "command": "task", "subcommand": "run", "args": ["--all"] }
```

#### task cycle

Full Task traversal (plan + run)

```
POST { "command": "task", "subcommand": "cycle" }
```

### Message

Message management commands

#### message send

Send a message to Claude

```
POST { "command": "message", "subcommand": "send", "args": ["telegram", "Hello Claude"] }
```

**Source**:
- `telegram`: Sent from Telegram
- `cli`: Sent from CLI (default)

#### message list

```
POST { "command": "message", "subcommand": "list", "flags": { "page": 1 } }
```

#### message get

```
POST { "command": "message", "subcommand": "get", "args": ["1"] }
```

#### message status

```
POST { "command": "message", "subcommand": "status" }
```

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

### Schedule

Schedule management commands

#### schedule add

```
POST { "command": "schedule", "subcommand": "add", "args": ["0 9 * * *", "Daily task", "--project", "myproject", "--once"] }
```

**Flags**:
- `--project <id>`: Specify project (optional)
- `--once`: Auto-disable after single execution

#### schedule list

```
POST { "command": "schedule", "subcommand": "list", "args": ["--all"] }
```

#### schedule get

```
POST { "command": "schedule", "subcommand": "get", "args": ["1"] }
```

#### schedule enable/disable

```
POST { "command": "schedule", "subcommand": "enable", "args": ["1"] }
POST { "command": "schedule", "subcommand": "disable", "args": ["1"] }
```

#### schedule delete

```
POST { "command": "schedule", "subcommand": "delete", "args": ["1", "yes"] }
```

#### schedule runs

Query execution history

```
POST { "command": "schedule", "subcommand": "runs", "args": ["1"], "flags": { "page": 1 } }
```

#### schedule run

Query a specific execution record

```
POST { "command": "schedule", "subcommand": "run", "args": ["run_id"] }
```

#### schedule set

```
POST { "command": "schedule", "subcommand": "set", "args": ["1", "project", "myproject"] }
```

### Status

Query system status

```
POST { "command": "status" }
GET /?args=status
```

**Response Data**:
```json
{
  "max": 3,
  "used": 1,
  "available": 2,
  "sessions": 1
}
```

## Error Codes

| HTTP Status | Description |
|-------------|-------------|
| 200 | Success |
| 400 | Bad request (success=false) |

## Pagination

List commands support pagination

**Flags**:
- `-p <page>`: Page number (starts from 1)
- `-n <size>`: Page size (default: 10)

**Response**:
```json
{
  "page": 1,
  "page_size": 10,
  "total_items": 25,
  "total_pages": 3,
  "has_next": true,
  "has_prev": false
}
```

## HTTP Timeout Settings

- **ReadTimeout**: 10 seconds
- **WriteTimeout**: 30 minutes (considering Claude execution time)
- **IdleTimeout**: 60 seconds
