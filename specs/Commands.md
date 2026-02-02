# TALOS Commands Reference

모든 TALOS 명령어의 상세 사용법

---

## 목차

0. [초기화](#초기화) (1개)
1. [Project 관리](#project-관리) (6개)
2. [Project 실행](#project-실행) (2개)
3. [Phase 관리](#phase-관리) (4개)
4. [Task 관리](#task-관리) (6개)
5. [Memo 관리](#memo-관리) (4개)

**총 23개 명령어**

---

## 초기화

### talos init

**설명**: 새 프로젝트 폴더와 TALOS 환경 초기화

**사용법**:
```bash
talos init <project-id> ["<project-description>"]
```

**인자**:
- `project-id` (필수): 프로젝트 ID (영문 소문자, 숫자, 하이픈, 언더스코어만 허용)
- `project-description` (선택): 프로젝트 설명

**동작**:
1. 현재 위치에 `<project-id>` 이름의 폴더 생성
2. 폴더 내 `CLAUDE.md` 파일 생성 (기본 템플릿)
3. 폴더 내 `.talos/` 디렉토리 생성
4. `.talos/db` SQLite 파일 생성 및 스키마 초기화
5. `projects` 테이블에 project id와 description 자동 입력

**생성되는 구조**:
```
<project-id>/
├── CLAUDE.md          # 프로젝트 설정 템플릿
└── .talos/
    └── db             # SQLite 데이터베이스
```

**응답**:
```json
{
  "success": true,
  "project_id": "blog-api",
  "path": "/path/to/blog-api",
  "message": "Project initialized successfully"
}
```

**에러**:
```json
{
  "success": false,
  "error": "Directory already exists: blog-api"
}
```

```json
{
  "success": false,
  "error": "Invalid project ID: Blog-API",
  "hint": "Use lowercase letters, numbers, hyphens, and underscores only"
}
```

**예시**:
```bash
# description 없이
talos init blog-api

# description 포함
talos init blog-api "Developer blogging platform with markdown support"

# 하이픈과 언더스코어 사용
talos init my_ecommerce-api "E-commerce REST API"
```

**CLAUDE.md 템플릿**:
```markdown
# <project-id>

## Description
<project-description>

## Tech Stack
- Backend:
- Frontend:
- Database:

## Commands
- `talos project set '<json>'` - 프로젝트 설정
- `talos required` - 필수 입력 확인
- `talos project plan` - 플래닝 시작
- `talos project start` - 실행 시작
```

---

## Project 관리

### talos project set

**설명**: 프로젝트 생성 또는 전체 업데이트 (싱글톤)

**사용법**:
```bash
talos project set '<json>'
```

**JSON 포맷**:
```json
{
  "name": "Blog Platform",
  "description": "Developer blogging platform",
  "context": {
    "project_name": "Blog Platform",
    "description": "Developer blogging platform with markdown",
    "target_users": "Tech bloggers and readers",
    "deadline": "2026-03-01",
    "constraints": "Must support 10k concurrent users"
  },
  "tech": {
    "backend": "FastAPI",
    "frontend": "React 18",
    "database": "PostgreSQL",
    "cache": "Redis",
    "auth_service": "Auth0",
    "deployment": "Docker + AWS ECS"
  },
  "design": {
    "architecture": "Microservices",
    "auth_method": "JWT with 1h expiry",
    "api_style": "RESTful",
    "db_schema_users": "id, email, password_hash, role, created_at",
    "caching_strategy": "Cache-aside pattern",
    "rate_limiting": "100 req/min per user"
  }
}
```

**필수 필드**:
- `name`: 프로젝트 이름
- `context.project_name`: 프로젝트명
- `context.description`: 프로젝트 설명
- `tech.backend`: 백엔드 프레임워크
- `tech.frontend`: 프론트엔드 프레임워크
- `tech.database`: 데이터베이스
- `design.architecture`: 아키텍처 패턴
- `design.auth_method`: 인증 방식
- `design.api_style`: API 스타일

**응답**:
```json
{
  "success": true,
  "project_id": "P001",
  "message": "Project created successfully"
}
```

**에러**:
```json
{
  "success": false,
  "error": "Missing required field: tech.backend",
  "required_fields": [
    "name",
    "context.project_name",
    "context.description",
    "tech.backend",
    "tech.frontend",
    "tech.database",
    "design.architecture",
    "design.auth_method",
    "design.api_style"
  ]
}
```

**예시**:
```bash
talos project set '{
  "name": "Blog Platform",
  "description": "A modern blogging platform",
  "context": {
    "project_name": "Blog Platform",
    "description": "Developer blogging with markdown",
    "target_users": "Tech bloggers"
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
}'
```

---

### talos project get

**설명**: 프로젝트 정보 조회

**사용법**:
```bash
talos project get
```

**응답**:
```json
{
  "project": {
    "id": "P001",
    "name": "Blog Platform",
    "description": "Developer blogging platform",
    "status": "active",
    "created_at": "2026-02-02T10:00:00Z"
  },
  "context": {
    "project_name": "Blog Platform",
    "description": "Developer blogging platform",
    "target_users": "Tech bloggers"
  },
  "tech": {
    "backend": "FastAPI",
    "frontend": "React",
    "database": "PostgreSQL"
  },
  "design": {
    "architecture": "Microservices",
    "auth_method": "JWT",
    "api_style": "RESTful"
  }
}
```

**에러**:
```json
{
  "success": false,
  "error": "No project found"
}
```

---

### talos context set

**설명**: Context 정보만 수정

**사용법**:
```bash
talos context set '<json>'
```

**JSON 포맷**:
```json
{
  "project_name": "Blog Platform",
  "description": "Developer blogging platform with markdown",
  "target_users": "Tech bloggers and readers",
  "deadline": "2026-03-01",
  "constraints": "Must support 10k concurrent users",
  "business_goal": "Monthly revenue $10k"
}
```

**필수 필드**:
- `project_name`
- `description`

**응답**:
```json
{
  "success": true,
  "message": "Context updated successfully"
}
```

**예시**:
```bash
talos context set '{
  "project_name": "Blog Platform",
  "description": "Developer blogging platform",
  "deadline": "2026-04-01"
}'
```

---

### talos tech set

**설명**: Tech 정보만 수정

**사용법**:
```bash
talos tech set '<json>'
```

**JSON 포맷**:
```json
{
  "backend": "FastAPI",
  "frontend": "React 18",
  "database": "PostgreSQL",
  "cache": "Redis",
  "queue": "Celery",
  "auth_service": "Auth0",
  "payment_service": "Stripe",
  "email_service": "SendGrid",
  "storage": "AWS S3",
  "deployment": "Docker + AWS ECS",
  "ci_cd": "GitHub Actions",
  "monitoring": "Datadog"
}
```

**필수 필드**:
- `backend`
- `frontend`
- `database`

**응답**:
```json
{
  "success": true,
  "message": "Tech updated successfully"
}
```

**예시**:
```bash
talos tech set '{
  "backend": "FastAPI",
  "frontend": "React",
  "database": "PostgreSQL",
  "cache": "Redis"
}'
```

---

### talos design set

**설명**: Design 정보만 수정

**사용법**:
```bash
talos design set '<json>'
```

**JSON 포맷**:
```json
{
  "architecture": "Microservices",
  "auth_method": "JWT with 1h expiry",
  "api_style": "RESTful",
  "db_schema_users": "id, email, password_hash, role, created_at",
  "db_schema_posts": "id, user_id, title, content, published_at",
  "caching_strategy": "Cache-aside pattern",
  "rate_limiting": "100 req/min per user",
  "error_handling": "Centralized error handler",
  "logging": "Structured JSON logs"
}
```

**필수 필드**:
- `architecture`
- `auth_method`
- `api_style`

**응답**:
```json
{
  "success": true,
  "message": "Design updated successfully"
}
```

**예시**:
```bash
talos design set '{
  "architecture": "Monolithic",
  "auth_method": "JWT",
  "api_style": "RESTful"
}'
```

---

### talos required

**설명**: 필수 입력 중 누락된 항목 확인

**사용법**:
```bash
talos required
```

**응답** (누락 시):
```json
{
  "ready": false,
  "missing_required": [
    {
      "field": "context.project_name",
      "prompt": "What is the project name?",
      "examples": ["Blog Platform", "E-commerce API"]
    },
    {
      "field": "tech.backend",
      "prompt": "What backend framework?",
      "options": ["FastAPI", "Django", "Flask", "Express"],
      "custom_allowed": true
    },
    {
      "field": "design.architecture",
      "prompt": "What architecture pattern?",
      "options": ["Monolithic", "Microservices", "Serverless"],
      "custom_allowed": true
    }
  ],
  "total_missing": 3,
  "message": "Please configure required settings"
}
```

**응답** (완료 시):
```json
{
  "ready": true,
  "message": "All required fields configured"
}
```

---

## Project 실행

### talos project plan

**설명**: 전체 프로젝트 플래닝 시작

**사용법**:
```bash
talos project plan
```

**동작**:
1. 필수 설정 확인
2. Phase 생성 프롬프트
3. 각 Phase별 Task 생성 프롬프트
4. MASTER_PLAN.md 생성

**응답**:
```json
{
  "success": true,
  "mode": "planning",
  "message": "Planning mode started. Create phases and tasks.",
  "next_steps": [
    "Create phases: talos phase create '<json>'",
    "Create tasks: talos task push '<json>'",
    "Write MASTER_PLAN.md"
  ]
}
```

**에러** (필수 누락):
```json
{
  "success": false,
  "ready": false,
  "missing_required": [...],
  "message": "Configure required settings first"
}
```

---

### talos project start

**설명**: 전체 프로젝트 실행 시작

**사용법**:
```bash
talos project start
```

**동작**:
1. pending task 자동 pop
2. Manifest 제공
3. Task 실행
4. 완료/실패 처리
5. 다음 task로 이동

**응답**:
```json
{
  "success": true,
  "mode": "execution",
  "message": "Execution mode started",
  "next_task": "T001"
}
```

**에러** (task 없음):
```json
{
  "success": false,
  "error": "No pending tasks",
  "message": "All tasks completed or no tasks created"
}
```

---

## Phase 관리

### talos phase create

**설명**: Phase 생성

**사용법**:
```bash
talos phase create '<json>'
```

**JSON 포맷**:
```json
{
  "project_id": "P001",
  "name": "UI Planning",
  "description": "User interface design and wireframing",
  "order_num": 1
}
```

**필수 필드**:
- `project_id`
- `name`
- `order_num`

**응답**:
```json
{
  "success": true,
  "phase_id": "PH001",
  "name": "UI Planning",
  "message": "Phase created successfully"
}
```

**예시**:
```bash
talos phase create '{
  "project_id": "P001",
  "name": "Development",
  "description": "Implementation phase",
  "order_num": 3
}'
```

---

### talos phase list

**설명**: Phase 목록 조회

**사용법**:
```bash
talos phase list
```

**응답**:
```json
{
  "phases": [
    {
      "id": "PH001",
      "name": "UI Planning",
      "description": "User interface design",
      "order_num": 1,
      "status": "done",
      "tasks_total": 5,
      "tasks_done": 5
    },
    {
      "id": "PH002",
      "name": "API Design",
      "order_num": 2,
      "status": "active",
      "tasks_total": 8,
      "tasks_done": 3
    },
    {
      "id": "PH003",
      "name": "Development",
      "order_num": 3,
      "status": "pending",
      "tasks_total": 20,
      "tasks_done": 0
    }
  ],
  "total": 3
}
```

---

### talos phase <phase-id> plan

**설명**: 특정 Phase 플래닝

**사용법**:
```bash
talos phase PH002 plan
```

**동작**:
1. Phase 확인
2. Task 생성 프롬프트
3. Phase plan 문서 생성

**응답**:
```json
{
  "success": true,
  "phase_id": "PH002",
  "phase_name": "API Design",
  "mode": "planning",
  "message": "Phase planning started. Create tasks for this phase."
}
```

**에러**:
```json
{
  "success": false,
  "error": "Phase not found: PH999"
}
```

---

### talos phase <phase-id> start

**설명**: 특정 Phase 실행 시작

**사용법**:
```bash
talos phase PH002 start
```

**동작**:
1. Phase의 pending task만 실행
2. Phase 완료 시 다음 Phase로 이동 안 함 (명시적 호출 필요)

**응답**:
```json
{
  "success": true,
  "phase_id": "PH002",
  "phase_name": "API Design",
  "mode": "execution",
  "next_task": "T010",
  "message": "Phase execution started"
}
```

**에러** (task 없음):
```json
{
  "success": false,
  "error": "No pending tasks in phase PH002"
}
```

---

## Task 관리

### talos task push

**설명**: Task 추가 (validation 포함)

**사용법**:
```bash
talos task push '<json>'
```

**JSON 포맷**:
```json
{
  "phase_id": "PH002",
  "parent_id": null,
  "title": "Design User Authentication API",
  "content": "Design endpoints for login, logout, register, password reset",
  "level": "node",
  "skill": "",
  "references": [
    "specs/auth-requirements.md",
    "specs/security-guidelines.md"
  ]
}
```

**필수 필드**:
- `phase_id`
- `title`
- `content`

**선택 필드**:
- `parent_id`: 부모 task ID (계층 구조)
- `level`: '', 'node', 'leaf' (기본: '')
- `skill`: Skill ID
- `references`: 참조 파일 경로 배열

**응답**:
```json
{
  "success": true,
  "task_id": "T042",
  "title": "Design User Authentication API",
  "message": "Task created successfully"
}
```

**에러** (validation 실패):
```json
{
  "success": false,
  "error": "Missing required field: title",
  "required_fields": ["phase_id", "title", "content"],
  "your_task": {...}
}
```

**예시**:
```bash
talos task push '{
  "phase_id": "PH002",
  "title": "Implement Login API",
  "content": "Create POST /api/auth/login endpoint with JWT",
  "level": "leaf",
  "references": ["specs/api-spec.md"]
}'
```

---

### talos task pop

**설명**: 다음 pending Task 조회 (Manifest 포함)

**사용법**:
```bash
talos task pop
```

**응답**:
```json
{
  "task": {
    "id": "T042",
    "phase_id": "PH002",
    "parent_id": null,
    "title": "Implement Auth API",
    "content": "Create JWT-based authentication endpoints:\n- POST /api/auth/login\n- POST /api/auth/logout\n- POST /api/auth/refresh",
    "level": "leaf",
    "skill": "",
    "references": ["specs/auth-spec.md"],
    "status": "pending",
    "created_at": "2026-02-02T10:00:00Z"
  },
  "manifest": {
    "context": {
      "project_name": "Blog Platform",
      "description": "Developer blogging platform",
      "target_users": "Tech bloggers"
    },
    "tech": {
      "backend": "FastAPI",
      "frontend": "React",
      "database": "PostgreSQL"
    },
    "design": {
      "architecture": "Microservices",
      "auth_method": "JWT",
      "api_style": "RESTful"
    },
    "state": {
      "current_project": "P001",
      "current_phase": "PH002",
      "current_task": "T042",
      "next_task": "T043"
    },
    "memos": [
      {
        "key": "jwt_security",
        "data": {
          "value": "Use httpOnly cookies for refresh tokens",
          "priority": 1,
          "summary": "JWT security best practice"
        }
      }
    ]
  }
}
```

**에러** (task 없음):
```json
{
  "success": false,
  "error": "No pending tasks",
  "message": "All tasks completed"
}
```

---

### talos task start

**설명**: Task 실행 시작 (pending → doing)

**사용법**:
```bash
talos task start <task_id>
```

**동작**:
1. status: pending → doing
2. started_at 기록
3. state 자동 업데이트

**응답**:
```json
{
  "success": true,
  "task_id": "T042",
  "status": "doing",
  "started_at": "2026-02-02T10:30:00Z",
  "message": "Task started"
}
```

**에러**:
```json
{
  "success": false,
  "error": "Task not found: T999"
}
```

```json
{
  "success": false,
  "error": "Task already started",
  "current_status": "doing"
}
```

**예시**:
```bash
talos task start T042
```

---

### talos task complete

**설명**: Task 완료 처리 (doing → done)

**사용법**:
```bash
talos task complete <task_id> '<json>'
```

**JSON 포맷**:
```json
{
  "result": "success",
  "notes": "Implemented login, logout, and refresh endpoints. All tests passing.",
  "duration": "2.5h",
  "files_created": [
    "src/api/auth.py",
    "tests/test_auth.py"
  ],
  "commits": ["a1b2c3d"]
}
```

**필수 필드**:
- `result`: 결과 설명

**응답**:
```json
{
  "success": true,
  "task_id": "T042",
  "status": "done",
  "completed_at": "2026-02-02T13:00:00Z",
  "message": "Task completed successfully"
}
```

**에러**:
```json
{
  "success": false,
  "error": "Task not in doing status",
  "current_status": "pending"
}
```

**예시**:
```bash
talos task complete T042 '{
  "result": "success",
  "notes": "All endpoints implemented and tested",
  "duration": "2h"
}'
```

---

### talos task fail

**설명**: Task 실패 처리 (doing → failed)

**사용법**:
```bash
talos task fail <task_id> '<json>'
```

**JSON 포맷**:
```json
{
  "error": "Database connection failed",
  "details": "Connection timeout after 30s. Database server unreachable.",
  "retry_possible": true,
  "blockers": [
    "Need database server credentials",
    "VPN connection required"
  ],
  "attempted_solutions": [
    "Checked connection string",
    "Verified firewall rules"
  ]
}
```

**필수 필드**:
- `error`: 에러 설명

**응답**:
```json
{
  "success": true,
  "task_id": "T042",
  "status": "failed",
  "failed_at": "2026-02-02T11:00:00Z",
  "message": "Task marked as failed"
}
```

**예시**:
```bash
talos task fail T042 '{
  "error": "API rate limit exceeded",
  "details": "External auth service rate limit reached",
  "retry_possible": true
}'
```

---

### talos task status

**설명**: 전체 Task 진행 상황 조회

**사용법**:
```bash
talos task status
```

**응답**:
```json
{
  "summary": {
    "total": 100,
    "pending": 45,
    "doing": 3,
    "done": 50,
    "failed": 2
  },
  "progress": 50,
  "current_phase": {
    "id": "PH002",
    "name": "API Design",
    "tasks": {
      "total": 8,
      "done": 3,
      "progress": 37.5
    }
  },
  "current_task": {
    "id": "T042",
    "title": "Implement Auth API",
    "status": "doing",
    "started_at": "2026-02-02T10:30:00Z",
    "duration": "0.5h"
  },
  "failed_tasks": [
    {
      "id": "T015",
      "title": "Setup database",
      "error": "Connection failed"
    },
    {
      "id": "T028",
      "title": "Deploy to staging",
      "error": "Build failed"
    }
  ]
}
```

---

## Memo 관리

### talos memo set

**설명**: Memo 저장

**사용법**:
```bash
# Project 전역
talos memo set <key> '<json>'

# Phase 귀속
talos memo set <phase_id>:<key> '<json>'

# Task 귀속
talos memo set <phase_id>:<task_id>:<key> '<json>'
```

**JSON 포맷**:
```json
{
  "value": "Use httpOnly cookies for refresh tokens to prevent XSS attacks",
  "priority": 1,
  "summary": "JWT security best practice",
  "tags": ["security", "jwt", "cookies"]
}
```

**필수 필드**:
- `value`: 메모 내용

**선택 필드**:
- `priority`: 1 (중요), 2 (보통), 3 (사소함) - 기본: 2
- `summary`: 간단한 요약
- `tags`: 태그 배열

**응답**:
```json
{
  "success": true,
  "scope": "project",
  "key": "jwt_security",
  "priority": 1,
  "message": "Memo saved successfully"
}
```

**예시**:
```bash
# Project 전역
talos memo set jwt_best_practice '{
  "value": "Always use httpOnly cookies",
  "priority": 1,
  "summary": "Security best practice"
}'

# Phase 귀속
talos memo set PH002:api_conventions '{
  "value": "Use plural nouns for REST resources",
  "priority": 1
}'

# Task 귀속
talos memo set PH002:T042:implementation_notes '{
  "value": "Used bcrypt with 12 rounds for password hashing",
  "priority": 2,
  "tags": ["security", "password"]
}'
```

---

### talos memo get

**설명**: Memo 조회

**사용법**:
```bash
# Project 전역
talos memo get <key>

# Phase 귀속
talos memo get <phase_id>:<key>

# Task 귀속
talos memo get <phase_id>:<task_id>:<key>
```

**응답**:
```json
{
  "scope": "project",
  "key": "jwt_security",
  "data": {
    "value": "Use httpOnly cookies for refresh tokens",
    "priority": 1,
    "summary": "JWT security best practice",
    "tags": ["security", "jwt"]
  },
  "created_at": "2026-02-02T10:00:00Z",
  "updated_at": "2026-02-02T10:00:00Z"
}
```

**에러**:
```json
{
  "success": false,
  "error": "Memo not found: jwt_security"
}
```

**예시**:
```bash
talos memo get jwt_best_practice
talos memo get PH002:api_conventions
talos memo get PH002:T042:implementation_notes
```

---

### talos memo list

**설명**: Memo 목록 조회

**사용법**:
```bash
# 전체
talos memo list

# Phase 메모만
talos memo list <phase_id>

# Task 메모만
talos memo list <phase_id>:<task_id>
```

**응답** (전체):
```json
{
  "memos": {
    "project": [
      {
        "key": "jwt_security",
        "priority": 1,
        "summary": "JWT security best practice",
        "created_at": "2026-02-02T10:00:00Z"
      },
      {
        "key": "postgres_optimization",
        "priority": 2,
        "summary": "Database indexing tips",
        "created_at": "2026-02-02T11:00:00Z"
      }
    ],
    "phase": {
      "PH002": [
        {
          "key": "api_conventions",
          "priority": 1,
          "summary": "REST naming conventions",
          "created_at": "2026-02-02T12:00:00Z"
        }
      ]
    },
    "task": {
      "PH002:T042": [
        {
          "key": "implementation_notes",
          "priority": 2,
          "summary": "Implementation details",
          "created_at": "2026-02-02T13:00:00Z"
        }
      ]
    }
  },
  "total": 4
}
```

**응답** (Phase):
```json
{
  "scope": "phase",
  "scope_id": "PH002",
  "memos": [
    {
      "key": "api_conventions",
      "priority": 1,
      "summary": "REST naming conventions"
    }
  ],
  "total": 1
}
```

**예시**:
```bash
talos memo list
talos memo list PH002
talos memo list PH002:T042
```

---

### talos memo del

**설명**: Memo 삭제

**사용법**:
```bash
# Project 전역
talos memo del <key>

# Phase 귀속
talos memo del <phase_id>:<key>

# Task 귀속
talos memo del <phase_id>:<task_id>:<key>
```

**응답**:
```json
{
  "success": true,
  "scope": "project",
  "key": "jwt_security",
  "message": "Memo deleted successfully"
}
```

**에러**:
```json
{
  "success": false,
  "error": "Memo not found: jwt_security"
}
```

**예시**:
```bash
talos memo del old_note
talos memo del PH001:deprecated
talos memo del PH002:T042:temp_note
```

---

## 명령어 요약

### 초기화 (1개)
```bash
talos init <id> ["<desc>"]      # 프로젝트 폴더 및 환경 초기화
```

### Project (6개)
```bash
talos project set '<json>'      # 프로젝트 생성/업데이트
talos project get               # 프로젝트 조회
talos context set '<json>'      # Context 수정
talos tech set '<json>'         # Tech 수정
talos design set '<json>'       # Design 수정
talos required                  # 필수 입력 확인
```

### Project 실행 (2개)
```bash
talos project plan              # 전체 플래닝
talos project start             # 전체 실행
```

### Phase (4개)
```bash
talos phase create '<json>'     # Phase 생성
talos phase list                # Phase 목록
talos phase <id> plan           # Phase 플래닝
talos phase <id> start          # Phase 실행
```

### Task (6개)
```bash
talos task push '<json>'        # Task 추가
talos task pop                  # Task 조회 (manifest 포함)
talos task start <id>           # Task 시작
talos task complete <id> '<json>'  # Task 완료
talos task fail <id> '<json>'      # Task 실패
talos task status               # 진행 상황
```

### Memo (4개)
```bash
talos memo set <key> '<json>'   # Memo 저장
talos memo get <key>            # Memo 조회
talos memo list [<scope>]       # Memo 목록
talos memo del <key>            # Memo 삭제
```

---

## 일반적인 워크플로우

### 0. 프로젝트 초기화
```bash
# 새 프로젝트 생성
talos init my-project "My awesome project"

# 생성된 폴더로 이동
cd my-project
```

### 1. 프로젝트 초기 설정
```bash
# 필수 항목 확인
talos required

# 전체 설정 (한 번에)
talos project set '{...}'

# 또는 개별 설정
talos context set '{...}'
talos tech set '{...}'
talos design set '{...}'
```

### 2. Planning
```bash
# 전체 프로젝트 플래닝
talos project plan

# Phase 생성
talos phase create '{...}'
talos phase create '{...}'

# Task 생성
talos task push '{...}'
talos task push '{...}'
```

### 3. Execution
```bash
# 전체 실행
talos project start

# 또는 Phase별 실행
talos phase PH001 start
talos phase PH002 start

# Task 처리
result = talos task pop
talos task start result.task.id
# ... 작업 ...
talos task complete result.task.id '{...}'
```

### 4. Memo 활용
```bash
# 중요한 발견 저장
talos memo set important_finding '{
  "value": "...",
  "priority": 1
}'

# 다음 pop 시 자동 제공됨
```

---

## 에러 처리

모든 명령어는 다음 형식으로 에러 반환:

```json
{
  "success": false,
  "error": "에러 메시지",
  "details": "상세 설명 (선택)",
  "code": "ERROR_CODE (선택)"
}
```

**일반적인 에러 코드**:
- `MISSING_REQUIRED`: 필수 필드 누락
- `INVALID_JSON`: JSON 파싱 실패
- `NOT_FOUND`: 리소스 없음
- `INVALID_STATUS`: 잘못된 상태 전이
- `VALIDATION_ERROR`: Validation 실패

---

**TALOS Commands Reference v1.0**