# TALOS - Task And LLM Operating System

## 개요

Claude Code를 위한 장시간 자동 실행 시스템

**목표**:
- 프로젝트 수동 세팅 자동화 (30-50분 절약)
- 6시간 이상 무인 작업 가능
- 컨텍스트 한계 극복 (작업 메모리 10배 확장)

**철학**:
- Claude Code가 자동화 스크립트 생성 → bash 실행
- TALOS는 워크플로우 + 메모리 관리
- 한 줄 명령으로 프로젝트 완성

---

## 기술 스택

- **Python + SQLite**: 의존성 없음, 고성능
- **파일**: `.talos/db` 하나로 모든 것 관리
- **성능**: 1000개 Task도 1ms

---

## 계층 구조

### project → phase → task

```
project: Blog Platform
├─ phase: UI Planning
│  ├─ task: Wireframes
│  └─ task: Design system
├─ phase: API Design
│  ├─ task: Endpoint spec
│  └─ task: DB schema
└─ phase: Development
   ├─ task: Auth API
   └─ task: Posts CRUD
```

**특징**:
- **project**: 프로젝트 전체
- **phase**: 작업 단계 (UI기획, API설계, 개발 등)
- **task**: 실제 실행 단위

---

## 데이터베이스 스키마

### projects
```sql
CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'active',
    created_at TEXT NOT NULL
);
```

### phases
```sql
CREATE TABLE phases (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    order_num INTEGER,
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'active', 'done')),
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id)
);
```

### tasks
```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    phase_id TEXT NOT NULL,
    parent_id TEXT DEFAULT NULL,
    status TEXT NOT NULL DEFAULT 'pending' 
        CHECK(status IN ('pending', 'doing', 'done', 'failed')),
    title TEXT NOT NULL,
    level TEXT DEFAULT ''
        CHECK(level IN ('', 'node', 'leaf')),
    skill TEXT DEFAULT '',
    "references" TEXT DEFAULT '[]',  -- JSON array
    content TEXT DEFAULT '',
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    failed_at TEXT,
    FOREIGN KEY (phase_id) REFERENCES phases(id),
    FOREIGN KEY (parent_id) REFERENCES tasks(id)
);
```

### context (싱글톤)
```sql
CREATE TABLE context (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    data TEXT NOT NULL,  -- JSON
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

**JSON 포맷**:
```json
{
  "project_name": "Blog Platform",
  "description": "Developer blogging platform",
  "target_users": "Tech bloggers",
  "deadline": "2026-03-01"
}
```

### tech (싱글톤)
```sql
CREATE TABLE tech (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    data TEXT NOT NULL,  -- JSON
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

**JSON 포맷**:
```json
{
  "backend": "FastAPI",
  "frontend": "React",
  "database": "PostgreSQL",
  "cache": "Redis",
  "deployment": "Docker + AWS"
}
```

### design (싱글톤)
```sql
CREATE TABLE design (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    data TEXT NOT NULL,  -- JSON
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

**JSON 포맷**:
```json
{
  "architecture": "Microservices",
  "auth_method": "JWT",
  "api_style": "RESTful",
  "db_schema_users": "id, email, password_hash, created_at",
  "caching_strategy": "Cache-aside"
}
```

### state (자동 관리)
```sql
CREATE TABLE state (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
```

**자동 저장 항목**:
- `current_project`: 현재 프로젝트 ID
- `current_phase`: 현재 phase ID
- `current_task`: 현재 task ID
- `next_task`: 다음 task ID

**관리**: Task 명령 실행 시 TALOS가 자동 업데이트

### memos
```sql
CREATE TABLE memos (
    scope TEXT NOT NULL,     -- 'project', 'phase', 'task'
    scope_id TEXT NOT NULL,  -- project_id, phase_id, task_id
    key TEXT NOT NULL,
    data TEXT NOT NULL,      -- JSON
    priority INTEGER DEFAULT 2
        CHECK(priority IN (1, 2, 3)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (scope, scope_id, key)
);
```

**영역**:
- `project`: 프로젝트 전역 메모
- `phase`: 특정 phase 메모
- `task`: 특정 task 메모

**Priority**:
- `1`: 중요 (manifest에 자동 포함)
- `2`: 보통
- `3`: 사소함

**JSON 포맷**:
```json
{
  "value": "실제 메모 내용",
  "summary": "간단한 요약 (선택)",
  "tags": ["tag1", "tag2"]
}
```

---

## 명령어 레퍼런스

### Project 관리
```bash
talos project '<json>' # 프로젝트 정보 입력. Talos는 클로드 코드 내에서 작동하므로 프로젝트는 싱글톤.
```

**JSON 포맷**:
```json
{
  "name": "Blog Platform",
  "description": "Developer blogging platform",
  "context":{
    "project_name": "Blog Platform",
    "description": "Developer blogging platform with markdown",
    "target_users": "Tech bloggers and readers",
    "deadline": "2026-03-01",
    "constraints": "Must support 10k concurrent users"
  },
  "tech":{
    "backend": "FastAPI",
    "frontend": "React 18",
    "database": "PostgreSQL",
    "cache": "Redis",
    "auth_service": "Auth0",
    "deployment": "Docker + AWS ECS"
  },
  "design":{
    "architecture": "Microservices",
    "auth_method": "JWT with 1h expiry",
    "api_style": "RESTful",
    "db_schema_users": "id, email, password_hash, role, created_at",
    "caching_strategy": "Cache-aside pattern",
    "rate_limiting": "100 req/min per user"
  }
}
```

### Project 실행
```bash
talos plan                      # 모든 Phase 플래닝 절차 시작
talos exec                      # 모든 Phase 실행 시작
talos all                       # plan + exec
```

### Phase 관리
```bash
talos phase create '<json>'  # Phase 등록
talos phase list             # Phase 목록 조회
talos phase <phase-id> plan  # Phase 하위 Task 플래닝 절차 시작
talos phase <phase-id> start # Phase 하위 Task 실행 시작
talos phase <phase-id> all   # plan + start
```

**JSON 포맷**:
```json
{
  "project_id": "P001",
  "name": "UI Planning",
  "description": "User interface design phase",
  "order_num": 1
}
```

### Task 관리
```bash
talos task push '<json>'               # Task 추가
talos task pop                         # 다음 pending Task (manifest 포함)
talos task start <task_id>             # pending → doing
talos task complete <task_id> '<json>' # doing → done
talos task fail <task_id> '<json>'     # doing → failed
talos task status                      # 진행 상황
```

**push JSON 포맷**:
```json
{
  "phase_id": "PH001",
  "parent_id": null,
  "title": "Setup project",
  "content": "Create initial structure",
  "level": "node",
  "skill": "",
  "references": ["specs/requirements.md"]
}
```

**complete JSON 포맷**:
```json
{
  "result": "success",
  "notes": "Completed successfully",
  "duration": "2.5h"
}
```

**fail JSON 포맷**:
```json
{
  "error": "Database connection failed",
  "details": "Connection timeout after 30s",
  "retry_possible": true
}
```

### Memo 관리
```bash
talos memo set '<json>'
talos memo get [phase]:[task]:<key>
talos memo list [phase]:[task]
talos memo del [phase]:[task]:<key>
```

**영역 지정**:
```bash
# Project 전역
talos memo get jwt_config

# Phase 귀속
talos memo get PH001:api_decisions

# Task 귀속
talos memo get PH001:T042:implementation_notes
```

**JSON 포맷**:
```json
{
  "phase": "PH001",
  "task": "T042",
  "key": "jwt_config",
  "value": "Use httpOnly cookies for refresh tokens",
  "priority": 1,
  "summary": "JWT security best practice",
  "tags": ["security", "jwt"]
}
```

**조회**:
```bash
# 전체
talos memo list

# Project 메모만
talos memo list

# Phase 메모만
talos memo list PH001

# Task 메모만
talos memo list PH001:T042
```

**특징**:
- 덮어쓰기 가능 (최신 값으로 업데이트)
- 한 번만 설정하면 됨
- Task 반환 시 자동 포함 (manifest)

### 유틸리티
```bash
talos required                  # 필수 입력 중 입력하지 않은 항목 안내.
```

---

## Manifest 자동 반환

### pop 명령 응답

`talos task pop` 실행 시 Task + Manifest 함께 반환

```json
{
  "task": {
    "id": "T042",
    "phase_id": "PH002",
    "title": "Implement Auth API",
    "content": "Create JWT-based authentication endpoints",
    "level": "leaf",
    "skill": "",
    "references": ["specs/auth-spec.md"],
    "status": "pending"
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
        "scope": "project",
        "scope_id": "P001",
        "key": "jwt_security",
        "data": {
          "value": "Use httpOnly cookies",
          "priority": 1,
          "summary": "JWT best practice"
        }
      },
      {
        "scope": "phase",
        "scope_id": "PH002",
        "key": "api_conventions",
        "data": {
          "value": "RESTful naming: plural nouns",
          "priority": 1
        }
      }
    ]
  }
}
```

**Manifest 포함 내용**:
1. `context`: 전체 context 데이터
2. `tech`: 전체 tech 데이터
3. `design`: 전체 design 데이터
4. `state`: 현재 state
5. `memos`: priority 1인 메모만

**장점**:
- Claude가 매번 조회할 필요 없음
- 컨텍스트 자동 제공
- 토큰 절약

---

## 필수 입력 시스템

### 필수 항목

**context** (필수):
- `project_name`
- `description`

**tech** (필수):
- `backend`
- `frontend`
- `database`

**design** (필수):
- `architecture`
- `auth_method`
- `api_style`

### 워크플로우

```
User: "talos plan tasks"
    ↓
Claude: talos required
    ↓
Talos: Check required
    ↓
Missing → Return required items
    ↓
Claude: Interactive collection
    - 옵션 제시
    - 사용자 선택
    ↓
Claude: talos project '<json>'
    ↓
User: "talos plan tasks" (재시도)
    ↓
Talos: Ready → Proceed
```

### 대화 예시

```
Claude: "프로젝트 설정이 필요합니다.

**1. 백엔드 프레임워크**
A) FastAPI - 빠르고 현대적
B) Django - 풀스택
C) Express - Node.js

추천: FastAPI"

User: "A"

Claude: [모든 필수 항목 수집 후]

talos context set '{
  "project_name": "Blog Platform",
  "description": "Developer blogging platform"
}'

talos tech set '{
  "backend": "FastAPI",
  "frontend": "React",
  "database": "PostgreSQL"
}'

talos design set '{
  "architecture": "Monolithic",
  "auth_method": "JWT",
  "api_style": "RESTful"
}'

"✅ 설정 완료! 이제 'talos plan tasks' 가능합니다."
```

---

## 워크플로우

### 1. 초기 설정

```bash
User: "블로그 만들어"
Claude: [토론 모드 - 요구사항 수집]

User: "talos plan tasks"
Talos: "필수 설정 누락"

Claude: [대화형 수집]
  Backend? Frontend? Database? ...
  
Claude: 
  talos context set '<json>'
  talos tech set '<json>'
  talos design set '<json>'
```

### 2. Planning

```bash
User: "talos plan tasks"
Talos: ✅ Ready

Claude:
  1. talos project-create '<json>'
  2. talos phase-create '<json>' (여러 번)
  3. talos push '<json>' (Task 생성)
  4. MASTER_PLAN.md 작성
```

### 3. Execution

```bash
User: "talos exec tasks"

Claude:
  while True:
      result = talos pop
      
      # result.manifest 사용
      context = result.manifest.context
      tech = result.manifest.tech
      memos = result.manifest.memos
      
      task = result.task
      
      # references 읽기
      for ref in task.references:
          view(ref)
      
      # 실행
      talos start task.id
      ... 작업 ...
      talos complete task.id '<json>'
      
      # Context 관리
      if context > 80%:
          talos save-context
          /clear
          talos load-context
```

### 4. Memo 활용

```bash
# 중요한 발견
talos memo set project:P001:jwt_best_practice '{
  "value": "Use httpOnly cookies for refresh tokens",
  "priority": 1,
  "summary": "Security best practice"
}'

# Phase별 메모
talos memo set phase:PH002:api_naming '{
  "value": "Use plural nouns for resources",
  "priority": 1
}'

# Task별 메모
talos memo set task:T042:implementation '{
  "value": "Used bcrypt with 12 rounds",
  "priority": 2
}'

# 다음 pop 시 priority 1은 자동 포함됨
```

---

## Task Status

```
pending → doing → done/failed
```

**전이**:
- `talos task start`: pending → doing
- `talos task complete`: doing → done
- `talos task fail`: doing → failed

**크래시 복구**:
- 크래시 시 status='doing'으로 남음
- 재시작 후 감지 → 재개 가능

---

## 제약사항

### Task
- `title`, `content` 필수
- `phase_id` 필수
- `level`: '', 'node', 'leaf'
- `references`: JSON 배열

### 필수 설정
- context: project_name, description
- tech: backend, frontend, database
- design: architecture, auth_method, api_style

### Memo
- 영역: project, phase, task
- Priority: 1 (중요), 2 (보통), 3 (사소함)
- JSON 포맷 필수

---

## 성능

| Tasks | JSON | SQLite |
|-------|------|--------|
| 100   | 10ms | 1ms |
| 1,000 | 150ms | 1ms |
| 10,000| 2.5s | 2ms |

---

## 설치 및 초기화

### 바이너리 설치
```bash
# Go로 빌드된 바이너리 설치
go install github.com/your/talos/cmd/talos@latest
```

### 프로젝트 초기화

```bash
talos init <project-id> ["<project-description>"]
```

**동작**:
1. 현재 위치에 `<project-id>` 폴더 생성
2. 폴더 내 `CLAUDE.md` 파일 생성 (기본 템플릿)
3. 폴더 내 `.talos/db` SQLite 파일 생성
4. projects 테이블에 project id와 description 자동 입력

**예시**:
```bash
# description 없이
talos init blog-api

# description 포함
talos init blog-api "Developer blogging platform with markdown support"
```

**생성되는 구조**:
```
blog-api/
├── CLAUDE.md          # 프로젝트 설정 템플릿
└── .talos/
    └── db             # SQLite 데이터베이스
```

---

## 핵심 가치

1. **단순함**: 명령어 최소화
2. **자동화**: Manifest 자동 반환, state 자동 관리
3. **영역 기반**: project/phase/task 메모 분리
4. **필수 입력**: 설정 없이 시작 불가
5. **효율성**: 한 번 조회로 모든 컨텍스트

**TALOS = 한 줄 명령으로 프로젝트 완성** 🚀