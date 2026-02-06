# Claribot API 문서

## 개요

Claribot은 CLI, Web UI, 텔레그램 통합을 위한 RESTful HTTP API를 제공합니다. 기본 포트는 `9847`입니다.

- **Base URL**: `http://127.0.0.1:9847`
- **Content-Type**: `application/json`
- **미들웨어**: CORS → Auth → Router

## 인증

원격(localhost가 아닌) 요청은 JWT 쿠키를 통한 인증이 필요합니다. Localhost 요청(`127.0.0.1`, `::1`)은 인증을 우회합니다.

### Auth 엔드포인트

Auth 엔드포인트는 인증 미들웨어에서 제외됩니다.

#### POST /api/auth/setup

초기 비밀번호 설정. 시스템 최초 구성 시 1회 호출됩니다.

**Request**:
```json
{
  "password": "your-password",
  "totp_code": "123456"
}
```

**Response** (첫 번째 호출 - QR 설정용 TOTP URI 반환):
```json
{
  "totp_uri": "otpauth://totp/Claribot?secret=..."
}
```

**Response** (두 번째 호출 - TOTP 코드로 설정 완료):
```json
{
  "success": true
}
```
성공 시 `auth_token` 쿠키를 설정합니다.

#### POST /api/auth/login

비밀번호와 TOTP 코드로 인증합니다.

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
성공 시 `auth_token` 쿠키를 설정합니다.

#### GET /api/auth/status

인증 상태를 확인합니다.

**Response**:
```json
{
  "setup_completed": true,
  "authenticated": true
}
```

#### POST /api/auth/logout

인증 쿠키를 삭제합니다.

**Response**:
```json
{
  "success": true
}
```

#### GET /api/auth/totp-setup

QR 코드 생성을 위한 TOTP 설정 URI를 조회합니다.

**Response**:
```json
{
  "totp_uri": "otpauth://totp/Claribot?secret=..."
}
```

## Response Format

모든 RESTful 엔드포인트는 다음 구조를 반환합니다:

```json
{
  "success": true,
  "message": "응답 메시지",
  "data": { ... },
  "needs_input": false,
  "prompt": "",
  "context": "",
  "error_type": ""
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
| error_type | string | 에러 유형 식별자 (예: `auth_error`) (선택) |

## 페이지네이션

모든 목록 엔드포인트는 쿼리 파라미터로 페이지네이션을 지원합니다:

```
GET /api/projects?page=2&page_size=20
GET /api/tasks?all=true
```

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| page | int | 1 | 페이지 번호 (1부터 시작) |
| page_size | int | 10 | 페이지당 항목 수 |
| all | bool | false | 전체 항목 조회 (페이지네이션 무시) |

**Response 래퍼**:
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

## 상태 & 사용량

### GET /api/status

Claude 세션, 순회 상태, Task 통계를 포함한 시스템 상태를 조회합니다.

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
    "failed": 1
  }
}
```

### GET /api/usage

Claude API 사용량 통계를 조회합니다.

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

실시간 사용량 데이터의 비동기 갱신을 트리거합니다.

**Response**: `202 Accepted`
```json
{
  "success": true,
  "message": "Usage refresh started"
}
```

---

## 프로젝트

### GET /api/projects

전체 프로젝트 목록을 조회합니다 (페이지네이션).

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

새 프로젝트를 생성합니다.

**Request**:
```json
{
  "id": "project-id",
  "path": "/path/to/project",
  "description": "프로젝트 설명"
}
```

### GET /api/projects/stats

전체 프로젝트별 Task 통계를 조회합니다.

**Response Data**:
```json
{
  "stats": [
    {
      "project_id": "myproject",
      "total": 20,
      "done": 15,
      "todo": 3,
      "planned": 1,
      "running": 1
    }
  ]
}
```

### GET /api/projects/{id}

특정 프로젝트 상세 정보를 조회합니다.

### PATCH /api/projects/{id}

프로젝트 필드를 수정합니다.

**Request** (field/value 형식):
```json
{
  "field": "parallel",
  "value": "2"
}
```

**Request** (직접 형식):
```json
{
  "description": "새 설명",
  "parallel": 2
}
```

**수정 가능 필드** (단일 필드 형식):
- `description`: 프로젝트 설명
- `parallel`: 동시 Claude 인스턴스 수
- `category`: 프로젝트 카테고리
- `pinned`: 프로젝트 고정/해제

### DELETE /api/projects/{id}

프로젝트를 삭제합니다.

### POST /api/projects/{id}/switch

프로젝트를 전환합니다 (활성 컨텍스트로 설정). 이후 프로젝트 범위의 모든 API 호출은 이 컨텍스트를 사용합니다.

---

## Task

Task 엔드포인트는 프로젝트 선택이 필요합니다.

### GET /api/tasks

Task 목록을 조회합니다 (트리뷰와 플랫 리스트 + 페이지네이션 지원).

**Query Parameters**:

| Parameter | Type | Description |
|-----------|------|-------------|
| tree | bool | 전체 Task 트리 구조로 반환 |
| parent_id | int | 부모 Task ID로 필터링 |
| page | int | 페이지 번호 |
| page_size | int | 페이지당 항목 수 |
| all | bool | 전체 항목 조회 |

### POST /api/tasks

새 Task를 생성합니다.

**Request**:
```json
{
  "title": "Task 제목",
  "parent_id": 1,
  "spec": "Task 명세"
}
```

**Response**: `201 Created`

### GET /api/tasks/{id}

특정 Task 상세 정보를 조회합니다 (spec, plan, report 포함).

### PATCH /api/tasks/{id}

Task 필드를 수정합니다.

**Request**:
```json
{
  "field": "status",
  "value": "done"
}
```

**수정 가능 필드**:
- `title`: 제목
- `spec`: 명세
- `plan`: 실행 계획
- `report`: 실행 보고서
- `status`: 상태 (`todo`, `planned`, `split`, `done`, `failed`)
- `priority`: 실행 순서 우선순위 (정수)

### DELETE /api/tasks/{id}

Task를 삭제합니다.

### POST /api/tasks/{id}/plan

Claude를 사용하여 단일 Task의 계획을 생성합니다.

### POST /api/tasks/{id}/run

Claude를 사용하여 단일 Task를 실행합니다.

### POST /api/tasks/plan-all

적격한 모든 Task의 계획을 생성합니다.

### POST /api/tasks/run-all

적격한 모든 Task를 실행합니다.

### POST /api/tasks/cycle

전체 Task 순회를 실행합니다 (plan + run).

### POST /api/tasks/stop

현재 실행 중인 Task 순회를 중단합니다.

---

## 메시지

### GET /api/messages

전체 메시지 목록을 조회합니다 (페이지네이션).

**Query Parameters**: `page`, `page_size`, `all`

### POST /api/messages

Claude에 메시지를 전송합니다.

**Request**:
```json
{
  "content": "메시지 내용",
  "source": "gui"
}
```

**Source 값**:
- `telegram`: 텔레그램에서 전송
- `cli`: CLI에서 전송
- `gui`: Web UI에서 전송
- `api`: 기본값 (source 미지정 시)

### GET /api/messages/{id}

특정 메시지 상세 정보를 조회합니다.

### GET /api/messages/status

메시지 큐 상태를 조회합니다.

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

현재 처리 중인 메시지를 조회합니다.

---

## Config

전역 설정 키-값 저장소.

### GET /api/configs

전체 설정 키-값 목록을 조회합니다.

### GET /api/configs/{key}

특정 설정 값을 조회합니다.

### PUT /api/configs/{key}

설정 값을 저장합니다.

**Request**:
```json
{
  "value": "config-value"
}
```

### DELETE /api/configs/{key}

설정 항목을 삭제합니다.

### GET /api/config-yaml

설정 YAML 파일 원본 내용을 조회합니다.

### PUT /api/config-yaml

설정 YAML 파일을 업데이트합니다.

**Request**:
```json
{
  "content": "key1: value1\nkey2: value2"
}
```

---

## 스케줄

### GET /api/schedules

스케줄 목록을 조회합니다 (현재 프로젝트 컨텍스트 기준).

**Query Parameters**:

| Parameter | Type | Description |
|-----------|------|-------------|
| all | bool | 전체 프로젝트의 스케줄 표시 |
| project_id | string | 프로젝트 ID로 필터링 |

### POST /api/schedules

새 스케줄을 생성합니다.

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
| cron_expr | string | 필수 | Cron 표현식 |
| message | string | 필수 | 실행할 명령/메시지 |
| type | string | `"claude"` | 실행 유형: `claude` (Claude Code) 또는 `bash` (쉘 명령) |
| project_id | string | 현재 | 대상 프로젝트 (기본값: 활성 프로젝트) |
| run_once | bool | false | 1회 실행 후 자동 비활성화 |

**Response**: `201 Created`

### GET /api/schedules/{id}

특정 스케줄 상세 정보를 조회합니다.

### PATCH /api/schedules/{id}

스케줄 필드를 수정합니다.

**Request**:
```json
{
  "field": "project",
  "value": "new-project-id"
}
```

**수정 가능 필드**:
- `project`: 스케줄을 프로젝트에 할당 (빈 문자열 또는 `"none"`으로 할당 해제)

### DELETE /api/schedules/{id}

스케줄을 삭제합니다.

### POST /api/schedules/{id}/enable

스케줄을 활성화합니다.

### POST /api/schedules/{id}/disable

스케줄을 비활성화합니다.

### GET /api/schedules/{id}/runs

스케줄의 실행 기록을 조회합니다 (페이지네이션).

### GET /api/schedule-runs/{runId}

특정 실행 기록을 조회합니다.

---

## Spec

프로젝트 요구사항 명세 관리. 프로젝트 선택이 필요합니다.

### GET /api/specs

전체 명세 목록을 조회합니다 (페이지네이션).

**Query Parameters**: `page`, `page_size`, `all`

### POST /api/specs

새 명세를 생성합니다.

**Request**:
```json
{
  "title": "명세 제목",
  "content": "명세 내용"
}
```

**Response**: `201 Created`

### GET /api/specs/{id}

특정 명세 상세 정보를 조회합니다.

### PATCH /api/specs/{id}

명세 필드를 수정합니다.

**Request**:
```json
{
  "field": "content",
  "value": "수정된 내용"
}
```

**수정 가능 필드**:
- `title`: 명세 제목
- `content`: 명세 내용
- `status`: 상태 (`draft`, `review`, `approved`, `deprecated`)
- `priority`: 우선순위 (정수)

### DELETE /api/specs/{id}

명세를 삭제합니다.

---

## Legacy API

하위 호환을 위해 기존 명령어 스타일 API도 사용 가능합니다:

```
GET /?args=project+list
POST /api { "command": "project", "subcommand": "list" }
GET /api/health
```

이 엔드포인트는 CLI (`clari`) 및 텔레그램 핸들러에서 사용됩니다.

---

## 에러 처리

| HTTP Status | Description |
|-------------|-------------|
| 200 | 성공 |
| 201 | 리소스 생성됨 |
| 202 | 비동기 작업 수락됨 |
| 400 | 잘못된 요청 (success=false) |
| 401 | 인증 필요 |
| 405 | 허용되지 않는 메서드 |
| 500 | 내부 서버 오류 |

## HTTP 타임아웃 설정

- **ReadTimeout**: 10초
- **WriteTimeout**: 30분 (Claude 실행 고려)
- **IdleTimeout**: 60초
