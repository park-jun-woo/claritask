# Claritask Commands Reference

> **버전**: v0.0.2

## 변경이력
| 버전 | 날짜 | 내용 |
|------|------|------|
| v0.0.2 | 2026-02-03 | Expert 명령어 추가 |
| v0.0.1 | 2026-02-03 | 최초 작성 |

## 개요

Claritask CLI의 모든 명령어 레퍼런스. 현재 구현 상태와 향후 계획을 구분하여 기술합니다.

**바이너리**: `clari`
**기술 스택**: Go + Cobra + SQLite

---

## 명령어 구조

```
clari
├── init                    # 프로젝트 초기화
├── project                 # 프로젝트 관리
│   ├── set / get / plan / start / stop / status
├── task                    # 작업 관리
│   ├── push / pop / start / complete / fail
│   ├── status / get / list
├── feature                 # Feature 관리
│   ├── list / add / get / spec / start
├── edge                    # Edge (의존성) 관리
│   ├── add / list / remove / infer
├── fdl                     # FDL 관리
│   ├── create / register / validate / show
│   ├── skeleton / tasks / verify / diff
├── plan                    # Planning 명령어
│   └── features
├── expert                  # Expert 관리
│   ├── add / list / get / edit / remove
│   ├── assign / unassign
├── memo                    # 메모 관리
│   ├── set / get / list / del
├── context                 # 컨텍스트 관리
│   ├── set / get
├── tech                    # 기술 스택 관리
│   ├── set / get
├── design                  # 설계 결정 관리
│   ├── set / get
└── required                # 필수 설정 확인
```

---

## 구현 상태

| 카테고리 | 명령어 수 | 상태 |
|----------|----------|------|
| 초기화 | 1 | 구현 완료 |
| Project | 6 | 구현 완료 |
| Task | 8 | 구현 완료 |
| Feature | 5 | 구현 완료 |
| Edge | 4 | 구현 완료 |
| FDL | 8 | 구현 완료 |
| Plan | 1 | 구현 완료 |
| Expert | 7 | 미구현 |
| Memo | 4 | 구현 완료 |
| Context/Tech/Design | 6 | 구현 완료 |
| Required | 1 | 구현 완료 |
| **총계** | **51** | - |

---

## 초기화

### clari init

프로젝트 초기화. LLM과 협업하여 프로젝트 설정 완성.

```bash
clari init <project-id> [options]
```

**인자**:
- `project-id` (필수): 프로젝트 ID
  - 규칙: 영문 소문자, 숫자, 하이픈(`-`), 언더스코어(`_`)만 허용

**옵션**:
| 옵션 | 단축 | 설명 |
|------|------|------|
| --name | -n | 프로젝트 이름 (기본값: project-id) |
| --description | -d | 프로젝트 설명 |
| --skip-analysis | | 컨텍스트 분석 건너뛰기 |
| --skip-specs | | Specs 생성 건너뛰기 |
| --non-interactive | | 비대화형 모드 (자동 승인) |
| --force | | 기존 DB 덮어쓰기 |
| --resume | | 중단된 초기화 재개 |

**프로세스**:
1. **Phase 1**: DB 초기화 (.claritask/db.clt 생성)
2. **Phase 2**: 프로젝트 파일 분석 (claude --print)
3. **Phase 3**: tech/design 승인 (대화형)
4. **Phase 4**: Specs 초안 생성 (claude --print)
5. **Phase 5**: 피드백 루프 (승인까지 반복)

**생성 구조**:
```
./
├── .claritask/
│   └── db
└── specs/
    └── <project-id>.md
```

**응답**:
```json
{
  "success": true,
  "project_id": "my-api",
  "db_path": ".claritask/db.clt",
  "specs_path": "specs/my-api.md"
}
```

**에러**:
```json
{
  "success": false,
  "error": "database already exists at .claritask/db.clt (use --force to overwrite)"
}
```

**예시**:
```bash
# 기본 사용 (전체 프로세스)
clari init my-api

# 옵션 지정
clari init my-api --name "My REST API" --description "사용자 관리 API"

# 빠른 초기화 (LLM 호출 없이)
clari init my-api --skip-analysis --skip-specs

# 기존 프로젝트 재초기화
clari init my-api --force

# 중단된 초기화 재개
clari init --resume

# 비대화형 모드 (CI/CD용)
clari init my-api --non-interactive
```

**Phase 상세**:

| Phase | 설명 | 건너뛰기 |
|-------|------|----------|
| 1 | .claritask/db.clt 생성, 프로젝트 레코드 | 불가 |
| 2 | 파일 스캔, LLM으로 tech/design 분석 | --skip-analysis |
| 3 | 분석 결과 사용자 승인 | --non-interactive (자동 승인) |
| 4 | LLM으로 specs 문서 생성 | --skip-specs |
| 5 | 피드백 반영, 최종 승인 | --non-interactive (자동 승인) |

---

## Project 관리

### clari project set

프로젝트 설정 일괄 입력 (context, tech, design)

```bash
clari project set '<json>'
```

**JSON 포맷**:
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

**응답**:
```json
{
  "success": true,
  "message": "Project settings updated successfully"
}
```

---

### clari project get

프로젝트 전체 정보 조회

```bash
clari project get
```

**응답**:
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

### clari project plan

플래닝 모드 시작 (필수 설정 확인)

```bash
clari project plan
```

**응답** (Ready):
```json
{
  "success": true,
  "ready": true,
  "project_id": "blog-api",
  "mode": "planning",
  "message": "Project is ready for planning. Use 'clari feature add' to add features."
}
```

**응답** (Not Ready):
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

### clari project start

실행 모드 시작 (Task 상태 확인)

```bash
clari project start
```

**응답**:
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

## Task 관리

### clari task push

새 Task 추가

```bash
clari task push '<json>'
```

**JSON 포맷**:
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

**필수 필드**: `feature_id`, `title`, `content`

**응답**:
```json
{
  "success": true,
  "task_id": 1,
  "title": "user_table_sql",
  "message": "Task created successfully"
}
```

---

### clari task pop

다음 실행할 Task 조회 (pending → doing 자동 전환)

```bash
clari task pop
```

**동작**:
1. `status = 'pending'`인 Task 중 가장 낮은 ID 선택
2. Task의 status를 `doing`으로 변경
3. `started_at` 기록
4. Manifest 구성 (context, tech, design, state, memos)
5. State 업데이트 (current_project, current_feature, current_task, next_task)

**응답**:
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

**응답** (Task 없음):
```json
{
  "success": true,
  "task": null,
  "message": "No pending tasks"
}
```

---

### clari task start

Task 수동 시작 (pending → doing)

```bash
clari task start <task_id>
```

**응답**:
```json
{
  "success": true,
  "task_id": 3,
  "status": "doing",
  "message": "Task started"
}
```

**에러**:
```json
{
  "success": false,
  "error": "task status must be 'pending' to start, current: doing"
}
```

---

### clari task complete

Task 완료 처리 (doing → done)

```bash
clari task complete <task_id> '<json>'
```

**JSON 포맷**:
```json
{
  "result": "JWT 기반 인증 서비스 구현 완료. login, logout, refresh 메서드 포함.",
  "notes": "토큰 만료: access 1h, refresh 7d"
}
```

**필수 필드**: `result`

**동작**:
1. status를 `done`으로 변경
2. `result` 저장 (의존 Task에 전달됨)
3. `completed_at` 기록

**응답**:
```json
{
  "success": true,
  "task_id": 3,
  "status": "done",
  "message": "Task completed"
}
```

---

### clari task fail

Task 실패 처리 (doing → failed)

```bash
clari task fail <task_id> '<json>'
```

**JSON 포맷**:
```json
{
  "error": "Database connection failed",
  "details": "Connection timeout after 30s"
}
```

**필수 필드**: `error`

**응답**:
```json
{
  "success": true,
  "task_id": 3,
  "status": "failed",
  "message": "Task failed"
}
```

---

### clari task status

전체 Task 진행 상황 조회

```bash
clari task status
```

**응답**:
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

### clari task get

특정 Task 상세 조회

```bash
clari task get <task_id>
```

**응답**:
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

### clari task list

Feature별 Task 목록 조회

```bash
clari task list [feature_id]
```

**응답**:
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

## Memo 관리

### clari memo set

Memo 저장 (upsert)

```bash
clari memo set <key> '<json>'
```

**Key 포맷**:
| 포맷 | Scope | 예시 |
|------|-------|------|
| `key` | project | `jwt_security` |
| `feature_id:key` | feature | `1:api_decisions` |
| `feature_id:task_id:key` | task | `1:3:blockers` |

**JSON 포맷**:
```json
{
  "value": "Use httpOnly cookies for refresh tokens",
  "priority": 1,
  "summary": "JWT 보안 정책",
  "tags": ["security", "jwt"]
}
```

**Priority**:
- `1`: 중요 - manifest에 자동 포함
- `2`: 보통 (기본값)
- `3`: 사소함

**응답**:
```json
{
  "success": true,
  "scope": "project",
  "scope_id": "",
  "key": "jwt_security",
  "message": "Memo saved successfully"
}
```

---

### clari memo get

Memo 조회

```bash
clari memo get <key>
```

**응답**:
```json
{
  "success": true,
  "scope": "project",
  "scope_id": "",
  "key": "jwt_security",
  "data": {
    "value": "Use httpOnly cookies",
    "summary": "JWT 보안 정책",
    "tags": ["security"]
  },
  "priority": 1,
  "created_at": "2026-02-03T10:00:00Z",
  "updated_at": "2026-02-03T10:00:00Z"
}
```

---

### clari memo list

Memo 목록 조회

```bash
clari memo list           # 전체
clari memo list <scope>   # Scope별
```

**응답** (전체):
```json
{
  "success": true,
  "memos": {
    "project": {
      "": [{"key": "jwt_security", "priority": 1, "summary": "JWT 보안 정책"}]
    },
    "feature": {
      "1": [{"key": "api_conventions", "priority": 1, "summary": "API 규칙"}]
    },
    "task": {
      "1:3": [{"key": "notes", "priority": 2, "summary": "구현 메모"}]
    }
  },
  "total": 3
}
```

---

### clari memo del

Memo 삭제

```bash
clari memo del <key>
```

**응답**:
```json
{
  "success": true,
  "message": "Memo deleted successfully"
}
```

---

## Context/Tech/Design 관리

### clari context set / get

프로젝트 컨텍스트 관리

```bash
clari context set '<json>'
clari context get
```

**필수 필드**: `project_name`, `description`

**JSON 포맷**:
```json
{
  "project_name": "Blog Platform",
  "description": "Developer blogging platform",
  "target_users": "Tech bloggers",
  "deadline": "2026-03-01"
}
```

---

### clari tech set / get

기술 스택 관리

```bash
clari tech set '<json>'
clari tech get
```

**필수 필드**: `backend`, `frontend`, `database`

**JSON 포맷**:
```json
{
  "backend": "FastAPI",
  "frontend": "React",
  "database": "PostgreSQL",
  "cache": "Redis",
  "deployment": "Docker"
}
```

---

### clari design set / get

설계 결정 관리

```bash
clari design set '<json>'
clari design get
```

**필수 필드**: `architecture`, `auth_method`, `api_style`

**JSON 포맷**:
```json
{
  "architecture": "Monolithic",
  "auth_method": "JWT",
  "api_style": "RESTful",
  "db_schema_users": "id, email, password_hash",
  "rate_limiting": "100 req/min"
}
```

---

### clari required

필수 설정 확인

```bash
clari required
```

**응답** (Ready):
```json
{
  "success": true,
  "ready": true,
  "message": "All required fields configured"
}
```

**응답** (Not Ready):
```json
{
  "success": true,
  "ready": false,
  "missing_required": [
    {"field": "context.project_name", "prompt": "프로젝트 이름을 입력하세요"},
    {"field": "tech.backend", "prompt": "백엔드 기술을 선택하세요", "options": ["go", "node", "python", "java"]},
    {"field": "design.architecture", "prompt": "아키텍처 패턴을 선택하세요", "options": ["monolith", "microservice", "serverless"]}
  ],
  "total_missing": 3,
  "message": "Please configure required settings"
}
```

---

## 워크플로우 예시

### 1. 프로젝트 초기화

```bash
# 프로젝트 생성
clari init blog-api "Developer blogging platform"
cd blog-api

# 필수 설정 확인
clari required

# 전체 설정
clari project set '{
  "name": "Blog Platform",
  "context": {"project_name": "Blog Platform", "description": "Developer blogging"},
  "tech": {"backend": "FastAPI", "frontend": "React", "database": "PostgreSQL"},
  "design": {"architecture": "Monolithic", "auth_method": "JWT", "api_style": "RESTful"}
}'
```

### 2. Planning

```bash
# 플래닝 모드 시작
clari project plan

# Feature 생성
clari feature add '{"name": "user_auth", "description": "사용자 인증 시스템"}'
clari feature add '{"name": "blog_posts", "description": "블로그 포스트 관리"}'

# Task 추가
clari task push '{"feature_id": 1, "title": "user_table_sql", "content": "CREATE TABLE users..."}'
clari task push '{"feature_id": 1, "title": "user_model", "content": "User 모델 구현"}'
```

### 3. Execution

```bash
# 실행 시작
clari project start

# Task 가져오기 (자동으로 doing 상태로 변경)
clari task pop

# 작업 수행 후 완료
clari task complete 1 '{"result": "users 테이블 생성 완료"}'

# 다음 Task
clari task pop
clari task complete 2 '{"result": "User 모델 구현 완료"}'

# 진행 상황 확인
clari task status
```

### 4. Memo 활용

```bash
# 중요한 발견 저장 (priority 1)
clari memo set jwt_security '{"value": "Use httpOnly cookies", "priority": 1}'

# 다음 task pop 시 manifest에 자동 포함됨
clari task pop
```

---

## 에러 처리

모든 명령어는 다음 형식으로 에러 반환:

```json
{
  "success": false,
  "error": "에러 메시지"
}
```

**일반적인 에러**:
- `open database: ...` - DB 연결 실패
- `parse JSON: ...` - JSON 파싱 실패
- `missing required field: ...` - 필수 필드 누락
- `task status must be '...' to ...` - 잘못된 상태 전이
- `memo not found` - 메모 없음

---

## Feature 관리

### clari feature list

Feature 목록 조회

```bash
clari feature list
```

**응답**:
```json
{
  "success": true,
  "features": [
    {
      "id": 1,
      "name": "user_auth",
      "spec": "사용자 인증 시스템",
      "status": "active",
      "fdl_hash": "abc123...",
      "skeleton_generated": true
    }
  ],
  "total": 1
}
```

---

### clari feature add

새 Feature 추가

```bash
clari feature add '<json>'
```

**JSON 포맷**:
```json
{
  "name": "user_auth",
  "description": "사용자 인증 시스템"
}
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "name": "user_auth",
  "message": "Feature created successfully"
}
```

---

### clari feature get

Feature 상세 조회

```bash
clari feature get <id>
```

**응답**:
```json
{
  "success": true,
  "feature": {
    "id": 1,
    "name": "user_auth",
    "spec": "사용자 인증 시스템",
    "fdl": "feature: user_auth\n...",
    "fdl_hash": "abc123...",
    "skeleton_generated": true
  }
}
```

---

### clari feature spec

Feature 스펙 설정

```bash
clari feature spec <id> '<spec_text>'
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "message": "Spec updated successfully"
}
```

---

### clari feature start

Feature 실행 시작

```bash
clari feature start <id>
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "status": "active",
  "message": "Feature started"
}
```

---

## Edge 관리

### clari edge add

의존성 Edge 추가

```bash
clari edge add --from <id> --to <id>
clari edge add --feature --from <id> --to <id>
```

**플래그**:
- `--from`: 의존하는 ID (필수)
- `--to`: 의존받는 ID (필수)
- `--feature`: Feature Edge로 추가 (기본값: Task Edge)

**응답**:
```json
{
  "success": true,
  "type": "task",
  "from_id": 2,
  "to_id": 1,
  "message": "Task edge created: user_model depends on user_table_sql"
}
```

---

### clari edge list

Edge 목록 조회

```bash
clari edge list
clari edge list --feature
clari edge list --task
clari edge list --feature-id <id>
```

**응답**:
```json
{
  "success": true,
  "feature_edges": [
    {
      "from": {"id": 2, "name": "comments"},
      "to": {"id": 1, "name": "user_auth"}
    }
  ],
  "task_edges": [
    {
      "from": {"id": "3", "title": "auth_service"},
      "to": {"id": "1", "title": "user_table_sql"}
    }
  ],
  "total_feature_edges": 1,
  "total_task_edges": 1
}
```

---

### clari edge remove

Edge 삭제

```bash
clari edge remove --from <id> --to <id>
clari edge remove --feature --from <id> --to <id>
```

**응답**:
```json
{
  "success": true,
  "type": "task",
  "from_id": 2,
  "to_id": 1,
  "message": "Edge removed successfully"
}
```

---

### clari edge infer

LLM 기반 Edge 추론

```bash
clari edge infer --feature <id>
clari edge infer --project
```

**플래그**:
- `--feature <id>`: Feature 내 Task Edge 추론
- `--project`: Feature 간 Edge 추론
- `--min-confidence`: 최소 확신도 (기본값: 0.7)

**응답**:
```json
{
  "success": true,
  "type": "task",
  "feature_id": 1,
  "items": [...],
  "existing_edges": [...],
  "prompt": "Analyze the following tasks...",
  "instructions": "Use the prompt to infer edges, then run 'clari edge add --from <id> --to <id>'"
}
```

---

## FDL 관리

### clari fdl create

FDL 템플릿 파일 생성

```bash
clari fdl create <name>
```

**응답**:
```json
{
  "success": true,
  "file": "features/user_auth.fdl.yaml",
  "message": "FDL template created. Edit the file and run 'clari fdl register'"
}
```

---

### clari fdl register

FDL 파일을 Feature로 등록

```bash
clari fdl register <file>
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "feature_name": "user_auth",
  "fdl_hash": "abc123...",
  "message": "FDL registered successfully"
}
```

---

### clari fdl validate

FDL 유효성 검증

```bash
clari fdl validate <feature_id>
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "valid": true,
  "message": "FDL is valid"
}
```

---

### clari fdl show

FDL 내용 조회

```bash
clari fdl show <feature_id>
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "feature_name": "user_auth",
  "fdl": "feature: user_auth\n...",
  "fdl_hash": "abc123...",
  "skeleton_generated": true
}
```

---

### clari fdl skeleton

FDL에서 스켈레톤 코드 생성

```bash
clari fdl skeleton <feature_id>
clari fdl skeleton <feature_id> --dry-run
clari fdl skeleton <feature_id> --force
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "generated_files": [
    {"path": "models/user.py", "layer": "model"},
    {"path": "services/user_auth_service.py", "layer": "service"}
  ],
  "total": 2,
  "message": "Skeletons generated successfully"
}
```

---

### clari fdl tasks

FDL에서 Task 자동 생성

```bash
clari fdl tasks <feature_id>
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "tasks_created": [
    {"id": 1, "title": "Create User model", "target_file": "models/user.py"}
  ],
  "edges_created": [
    {"from": 2, "to": 1}
  ],
  "total_tasks": 5,
  "total_edges": 3,
  "message": "Tasks generated from FDL"
}
```

---

### clari fdl verify

구현이 FDL과 일치하는지 검증

```bash
clari fdl verify <feature_id>
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "valid": false,
  "errors": ["Found 2 verification issues"],
  "functions_missing": ["createUser"],
  "files_missing": ["models/user.py"]
}
```

---

### clari fdl diff

FDL과 실제 코드 차이점 표시

```bash
clari fdl diff <feature_id>
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "feature_name": "user_auth",
  "differences": [
    {
      "file_path": "services/user_auth_service.py",
      "layer": "service",
      "changes": [
        {"type": "missing", "element": "function", "expected": "createUser"}
      ]
    }
  ],
  "total_changes": 1
}
```

---

## Plan 관리

### clari plan features

프로젝트 설명 기반 Feature 목록 생성

```bash
clari plan features
clari plan features --auto-create
```

**플래그**:
- `--auto-create`: 추론된 Feature 자동 생성

**응답**:
```json
{
  "success": true,
  "prompt": "You are analyzing a software project...",
  "instructions": "Use the prompt to generate features, then run 'clari feature add'"
}
```

---

## Expert 관리

Expert는 프로젝트에서 사용하는 전문가 역할 정의입니다. 기술 스택별 코딩 규칙, 아키텍처 패턴, 테스트 규칙 등을 정의합니다.

**저장 구조**:
```
.claritask/
└── experts/
    ├── backend-go-gin/
    │   └── EXPERT.md
    ├── frontend-react/
    │   └── EXPERT.md
    └── devops-k8s/
        └── EXPERT.md
```

---

### clari expert add

새 Expert 생성 (폴더 + 템플릿 파일)

```bash
clari expert add <expert-id>
```

**인자**:
- `expert-id` (필수): Expert ID
  - 규칙: 영문 소문자, 숫자, 하이픈(`-`)만 허용

**동작**:
1. `.claritask/experts/<expert-id>/` 폴더 생성
2. `EXPERT.md` 템플릿 파일 생성
3. (옵션) 에디터로 파일 열기

**응답**:
```json
{
  "success": true,
  "expert_id": "backend-go-gin",
  "path": ".claritask/experts/backend-go-gin/EXPERT.md",
  "message": "Expert created. Edit the file to define the expert."
}
```

**에러**:
```json
{
  "success": false,
  "error": "expert 'backend-go-gin' already exists"
}
```

---

### clari expert list

Expert 목록 조회

```bash
clari expert list
clari expert list --assigned          # 현재 프로젝트에 할당된 것만
clari expert list --available         # 할당 가능한 것만
```

**플래그**:
- `--assigned`: 현재 프로젝트에 할당된 Expert만 표시
- `--available`: 할당되지 않은 Expert만 표시

**응답**:
```json
{
  "success": true,
  "experts": [
    {
      "id": "backend-go-gin",
      "name": "Backend Go GIN Developer",
      "domain": "Backend API Development",
      "assigned": true
    },
    {
      "id": "frontend-react",
      "name": "Frontend React Developer",
      "domain": "Frontend Development",
      "assigned": false
    }
  ],
  "total": 2
}
```

---

### clari expert get

Expert 상세 정보 조회

```bash
clari expert get <expert-id>
```

**응답**:
```json
{
  "success": true,
  "expert": {
    "id": "backend-go-gin",
    "name": "Backend Go GIN Developer",
    "version": "1.0.0",
    "domain": "Backend API Development",
    "language": "Go 1.21+",
    "framework": "GIN Web Framework",
    "path": ".claritask/experts/backend-go-gin/EXPERT.md",
    "assigned": true
  }
}
```

**에러**:
```json
{
  "success": false,
  "error": "expert 'backend-go-gin' not found"
}
```

---

### clari expert edit

Expert 파일 편집 (에디터 실행)

```bash
clari expert edit <expert-id>
```

**동작**:
1. `$EDITOR` 환경변수로 지정된 에디터 실행
2. 없으면 `vi` 또는 `notepad` (OS에 따라)

**응답**:
```json
{
  "success": true,
  "expert_id": "backend-go-gin",
  "path": ".claritask/experts/backend-go-gin/EXPERT.md",
  "message": "Opening editor..."
}
```

---

### clari expert remove

Expert 삭제

```bash
clari expert remove <expert-id>
clari expert remove <expert-id> --force
```

**플래그**:
- `--force`: 확인 없이 삭제

**동작**:
1. 프로젝트에서 할당 해제
2. Expert 폴더 전체 삭제

**응답**:
```json
{
  "success": true,
  "expert_id": "backend-go-gin",
  "message": "Expert removed successfully"
}
```

**에러**:
```json
{
  "success": false,
  "error": "expert 'backend-go-gin' is assigned to project. Use --force to remove"
}
```

---

### clari expert assign

프로젝트에 Expert 할당

```bash
clari expert assign <expert-id>
clari expert assign <expert-id> --project <project-id>
```

**플래그**:
- `--project <project-id>`: 특정 프로젝트에 할당 (기본값: 현재 프로젝트)

**동작**:
1. DB의 `project_experts` 테이블에 관계 추가
2. Task pop 시 해당 Expert 내용이 manifest에 포함됨

**응답**:
```json
{
  "success": true,
  "expert_id": "backend-go-gin",
  "project_id": "my-api",
  "message": "Expert assigned to project"
}
```

**에러**:
```json
{
  "success": false,
  "error": "expert 'backend-go-gin' is already assigned to project 'my-api'"
}
```

---

### clari expert unassign

프로젝트에서 Expert 할당 해제

```bash
clari expert unassign <expert-id>
clari expert unassign <expert-id> --project <project-id>
```

**응답**:
```json
{
  "success": true,
  "expert_id": "backend-go-gin",
  "project_id": "my-api",
  "message": "Expert unassigned from project"
}
```

---

## Expert 템플릿

`clari expert add` 실행 시 생성되는 기본 템플릿:

```markdown
# Expert: [Expert Name]

## Metadata

| Field       | Value                          |
|-------------|--------------------------------|
| ID          | `expert-id`                    |
| Name        | Expert Name                    |
| Version     | 1.0.0                          |
| Domain      | Domain Description             |
| Language    | Language Version               |
| Framework   | Framework Name                 |

## Role Definition

[전문가 역할 설명 - 한 문장]

## Tech Stack

### Core
- **Language**:
- **Framework**:
- **Database**:

### Supporting
- **Auth**:
- **Validation**:
- **Logging**:
- **Testing**:

## Architecture Pattern

[디렉토리 구조]

## Coding Rules

[패턴별 코드 템플릿]

## Error Handling

[에러 처리 규칙]

## Testing Rules

[테스트 코드 규칙]

## Security Checklist

- [ ] 보안 항목들

## References

- [문서 링크]
```

---

## Expert와 Task Manifest 연동

프로젝트에 Expert가 할당되면, `clari task pop` 응답의 manifest에 포함됩니다:

```json
{
  "success": true,
  "task": {...},
  "manifest": {
    "context": {...},
    "tech": {...},
    "design": {...},
    "experts": [
      {
        "id": "backend-go-gin",
        "name": "Backend Go GIN Developer",
        "content": "# Expert: Backend Go GIN Developer\n..."
      }
    ],
    "state": {...},
    "memos": [...]
  }
}
```

---

## 확장된 Project 명령어

### clari project start (확장)

자동 실행 모드 옵션 지원

```bash
clari project start
clari project start --feature <id>
clari project start --dry-run
clari project start --fallback-interactive
```

**플래그**:
- `--feature <id>`: 특정 Feature만 실행
- `--dry-run`: 실행 계획만 표시
- `--fallback-interactive`: 실패 시 대화형 전환

---

### clari project stop

프로젝트 실행 중지

```bash
clari project stop
```

**응답**:
```json
{
  "success": true,
  "message": "Project execution stopped"
}
```

---

### clari project status

프로젝트 실행 상태 조회

```bash
clari project status
```

**응답**:
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

## 미구현 명령어 (향후 계획)

다음 명령어들은 향후 구현 예정입니다:

```bash
# FDL 심층 검증
clari fdl verify <feature_id> --strict    # 엄격한 검증 모드

# 자동 Edge 추가
clari edge infer --feature <id> --auto-add
```

---

*Claritask Commands Reference v3.2 - 2026-02-03*
