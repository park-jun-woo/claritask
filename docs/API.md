# Claribot API Documentation

## Overview

Claribot은 HTTP API를 통해 CLI와 통신합니다. 기본 포트는 `9847`입니다.

- **Base URL**: `http://127.0.0.1:9847`
- **Content-Type**: `application/json`

## Request Format

### POST (권장)

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

### GET (하위 호환)

```
GET /?args=project+list
```

## Response Format

```json
{
  "success": true,
  "message": "응답 메시지",
  "data": { ... },
  "needs_input": false,
  "prompt": "",
  "context": ""
}
```

| Field | Type | Description |
|-------|------|-------------|
| success | bool | 성공 여부 |
| message | string | 사용자에게 표시할 메시지 |
| data | any | 구조화된 응답 데이터 (선택) |
| needs_input | bool | 추가 입력 필요 여부 |
| prompt | string | 입력 프롬프트 (needs_input=true 시) |
| context | string | 다음 요청에 포함할 컨텍스트 |

## Commands

### Project

프로젝트 관리 명령어

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

작업 관리 명령어. 프로젝트 선택 필요.

#### task list

```
POST { "command": "task", "subcommand": "list", "flags": { "page": 1 } }
```

**Query Parameters**:
- `parent_id` (int, optional): 부모 Task ID
- `-p` (int): 페이지 번호
- `-n` (int): 페이지 크기

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
- `title`: 제목
- `status`: 상태 (todo, in_progress, done)
- `description`: 설명

#### task delete

```
POST { "command": "task", "subcommand": "delete", "args": ["1", "yes"] }
```

#### task plan

Claude를 사용하여 Task 계획 생성

```
POST { "command": "task", "subcommand": "plan", "args": ["1"] }
POST { "command": "task", "subcommand": "plan", "args": ["--all"] }
```

#### task run

Claude를 사용하여 Task 실행

```
POST { "command": "task", "subcommand": "run", "args": ["1"] }
POST { "command": "task", "subcommand": "run", "args": ["--all"] }
```

#### task cycle

전체 Task 순회 (plan + run)

```
POST { "command": "task", "subcommand": "cycle" }
```

### Message

메시지 관리 명령어

#### message send

Claude에 메시지 전송

```
POST { "command": "message", "subcommand": "send", "args": ["telegram", "Hello Claude"] }
```

**Source**:
- `telegram`: 텔레그램에서 전송
- `cli`: CLI에서 전송 (기본값)

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

스케줄 관리 명령어

#### schedule add

```
POST { "command": "schedule", "subcommand": "add", "args": ["0 9 * * *", "Daily task", "--project", "myproject", "--once"] }
```

**Flags**:
- `--project <id>`: 프로젝트 지정 (선택)
- `--once`: 1회 실행 후 자동 비활성화

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

실행 기록 조회

```
POST { "command": "schedule", "subcommand": "runs", "args": ["1"], "flags": { "page": 1 } }
```

#### schedule run

특정 실행 기록 조회

```
POST { "command": "schedule", "subcommand": "run", "args": ["run_id"] }
```

#### schedule set

```
POST { "command": "schedule", "subcommand": "set", "args": ["1", "project", "myproject"] }
```

### Status

시스템 상태 조회

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
| 200 | 성공 |
| 400 | 잘못된 요청 (success=false) |

## Pagination

리스트 명령어에서 페이지네이션 지원

**Flags**:
- `-p <page>`: 페이지 번호 (1부터 시작)
- `-n <size>`: 페이지 크기 (기본값: 10)

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

- **ReadTimeout**: 10초
- **WriteTimeout**: 30분 (Claude 실행 고려)
- **IdleTimeout**: 60초
