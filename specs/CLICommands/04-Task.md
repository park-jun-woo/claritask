# clari task - 작업 관리

> **버전**: v0.0.3

## clari task push

새 Task 추가

```bash
clari task push '<json>'
```

**JSON 포맷:**
```json
{
  "feature_id": 1,
  "title": "user_table_sql",
  "content": "CREATE TABLE users (id, email, password_hash, created_at)",
  "level": "leaf",
  "skill": "sql",
  "references": ["docs/schema.md"]
}
```

**필수 필드:** `feature_id`, `title`, `content`

**응답:**
```json
{
  "success": true,
  "task_id": 1,
  "title": "user_table_sql",
  "message": "Task created successfully"
}
```

---

## clari task pop

다음 실행할 Task 조회 (pending → doing 자동 전환)

```bash
clari task pop
```

**동작:**
1. `status = 'pending'`인 Task 중 가장 낮은 ID 선택
2. Task의 status를 `doing`으로 변경
3. `started_at` 기록
4. Manifest 구성 (context, tech, design, state, memos)
5. State 업데이트 (current_project, current_feature, current_task, next_task)

**응답:**
```json
{
  "success": true,
  "task": {
    "id": "3",
    "feature_id": "1",
    "parent_id": null,
    "status": "doing",
    "title": "auth_service",
    "level": "leaf",
    "skill": "python",
    "references": ["services/user.py"],
    "content": "JWT 기반 인증 서비스 구현",
    "result": "",
    "error": "",
    "created_at": "2026-02-03T10:00:00Z",
    "started_at": "2026-02-03T10:30:00Z"
  },
  "manifest": {
    "context": {
      "project_name": "Blog Platform",
      "description": "Developer blogging platform"
    },
    "tech": {
      "backend": "FastAPI",
      "frontend": "React",
      "database": "PostgreSQL"
    },
    "design": {
      "architecture": "Monolithic",
      "auth_method": "JWT",
      "api_style": "RESTful"
    },
    "state": {
      "current_project": "blog-api",
      "current_feature": "1",
      "current_task": "3",
      "next_task": "4"
    },
    "memos": [
      {
        "scope": "project",
        "scope_id": "",
        "key": "jwt_security",
        "data": {"value": "Use httpOnly cookies"},
        "priority": 1
      }
    ]
  }
}
```

**응답 (Task 없음):**
```json
{
  "success": true,
  "task": null,
  "message": "No pending tasks"
}
```

---

## clari task start

Task 수동 시작 (pending → doing)

```bash
clari task start <task_id>
```

**응답:**
```json
{
  "success": true,
  "task_id": 3,
  "status": "doing",
  "message": "Task started"
}
```

**에러:**
```json
{
  "success": false,
  "error": "task status must be 'pending' to start, current: doing"
}
```

---

## clari task complete

Task 완료 처리 (doing → done)

```bash
clari task complete <task_id> '<json>'
```

**JSON 포맷:**
```json
{
  "result": "JWT 기반 인증 서비스 구현 완료. login, logout, refresh 메서드 포함.",
  "notes": "토큰 만료: access 1h, refresh 7d"
}
```

**필수 필드:** `result`

**동작:**
1. status를 `done`으로 변경
2. `result` 저장 (의존 Task에 전달됨)
3. `completed_at` 기록

**응답:**
```json
{
  "success": true,
  "task_id": 3,
  "status": "done",
  "message": "Task completed"
}
```

---

## clari task fail

Task 실패 처리 (doing → failed)

```bash
clari task fail <task_id> '<json>'
```

**JSON 포맷:**
```json
{
  "error": "Database connection failed",
  "details": "Connection timeout after 30s"
}
```

**필수 필드:** `error`

**응답:**
```json
{
  "success": true,
  "task_id": 3,
  "status": "failed",
  "message": "Task failed"
}
```

---

## clari task status

전체 Task 진행 상황 조회

```bash
clari task status
```

**응답:**
```json
{
  "success": true,
  "summary": {
    "total": 42,
    "pending": 25,
    "doing": 1,
    "done": 15,
    "failed": 1,
    "progress": 35.7
  },
  "state": {
    "current_project": "blog-api",
    "current_feature": "2",
    "current_task": "16",
    "next_task": "17"
  },
  "progress": "35.7%"
}
```

---

## clari task get

특정 Task 상세 조회

```bash
clari task get <task_id>
```

**응답:**
```json
{
  "success": true,
  "task": {
    "id": "3",
    "feature_id": "1",
    "parent_id": null,
    "status": "done",
    "title": "auth_service",
    "level": "leaf",
    "skill": "python",
    "references": ["services/user.py"],
    "content": "JWT 기반 인증 서비스 구현",
    "result": "구현 완료. login, logout, refresh 포함.",
    "error": "",
    "created_at": "2026-02-03T10:00:00Z",
    "started_at": "2026-02-03T10:30:00Z",
    "completed_at": "2026-02-03T12:00:00Z"
  }
}
```

---

## clari task list

Feature별 Task 목록 조회

```bash
clari task list [feature_id]
```

**응답:**
```json
{
  "success": true,
  "tasks": [
    {"id": "1", "feature_id": "1", "status": "done", "title": "user_table_sql", ...},
    {"id": "2", "feature_id": "1", "status": "done", "title": "user_model", ...},
    {"id": "3", "feature_id": "1", "status": "doing", "title": "auth_service", ...}
  ],
  "total": 3
}
```

---

*Claritask Commands Reference v0.0.3 - 2026-02-03*
