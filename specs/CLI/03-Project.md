# clari project - 프로젝트 관리

> **버전**: v0.0.3

## clari project set

프로젝트 설정 일괄 입력 (context, tech, design)

```bash
clari project set '<json>'
```

**JSON 포맷:**
```json
{
  "context": {
    "project_name": "Blog Platform",
    "description": "Developer blogging with markdown",
    "target_users": "Tech bloggers",
    "deadline": "2026-03-01"
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
  }
}
```

**응답:**
```json
{
  "success": true,
  "message": "Project settings updated successfully"
}
```

---

## clari project get

프로젝트 전체 정보 조회

```bash
clari project get
```

**응답:**
```json
{
  "success": true,
  "project": {
    "id": "blog-api",
    "name": "Blog Platform",
    "description": "Developer blogging platform",
    "status": "active",
    "created_at": "2026-02-03T10:00:00Z"
  },
  "context": { "project_name": "...", "description": "..." },
  "tech": { "backend": "FastAPI", "frontend": "React", "database": "PostgreSQL" },
  "design": { "architecture": "Monolithic", "auth_method": "JWT", "api_style": "RESTful" }
}
```

---

## clari project plan

플래닝 모드 시작 (필수 설정 확인)

```bash
clari project plan
```

**응답 (Ready):**
```json
{
  "success": true,
  "ready": true,
  "project_id": "blog-api",
  "mode": "planning",
  "message": "Project is ready for planning. Use 'clari feature add' to add features."
}
```

**응답 (Not Ready):**
```json
{
  "success": false,
  "ready": false,
  "missing_required": [
    {"field": "context.project_name", "prompt": "프로젝트 이름을 입력하세요"},
    {"field": "tech.backend", "prompt": "백엔드 기술을 선택하세요", "options": ["go", "node", "python", "java"]}
  ],
  "message": "Please configure required settings before planning"
}
```

---

## clari project start

실행 모드 시작 (Task 상태 확인)

```bash
clari project start
clari project start --feature <id>
clari project start --dry-run
clari project start --fallback-interactive
```

**플래그:**
- `--feature <id>`: 특정 Feature만 실행
- `--dry-run`: 실행 계획만 표시
- `--fallback-interactive`: 실패 시 대화형 전환

**응답:**
```json
{
  "success": true,
  "ready": true,
  "mode": "execution",
  "status": {
    "total": 42,
    "pending": 25,
    "doing": 1,
    "done": 15,
    "failed": 1
  },
  "progress": 35.7,
  "message": "Project is ready for execution. Use 'clari task pop' to get the next task."
}
```

---

## clari project stop

프로젝트 실행 중지

```bash
clari project stop
```

**응답:**
```json
{
  "success": true,
  "message": "Project execution stopped"
}
```

---

## clari project status

프로젝트 실행 상태 조회

```bash
clari project status
```

**응답:**
```json
{
  "success": true,
  "project_id": "blog-api",
  "running": true,
  "progress": {
    "total": 42,
    "done": 15,
    "doing": 1,
    "pending": 25,
    "failed": 1
  },
  "current_task": {...}
}
```

---

*Claritask Commands Reference v0.0.3 - 2026-02-03*
