# Database: Settings Tables

> **현재 버전**: v0.0.4 ([변경이력](../HISTORY.md))

---

## context

프로젝트 컨텍스트 (싱글톤)

```sql
CREATE TABLE context (
    id INTEGER PRIMARY KEY CHECK(id = 1),  -- 싱글톤 보장
    data TEXT NOT NULL,                     -- JSON
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

**JSON 포맷**:
```json
{
  "project_name": "Blog Platform",
  "description": "Developer blogging platform with markdown",
  "target_users": "Tech bloggers and readers",
  "deadline": "2026-03-01",
  "constraints": "Must support 10k concurrent users"
}
```

**필수 필드**:
- `project_name`
- `description`

---

## tech

기술 스택 (싱글톤)

```sql
CREATE TABLE tech (
    id INTEGER PRIMARY KEY CHECK(id = 1),  -- 싱글톤 보장
    data TEXT NOT NULL,                     -- JSON
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

**JSON 포맷**:
```json
{
  "backend": "FastAPI",
  "frontend": "React 18",
  "database": "PostgreSQL",
  "cache": "Redis",
  "auth_service": "Auth0",
  "deployment": "Docker + AWS ECS"
}
```

**필수 필드**:
- `backend`
- `frontend`
- `database`

---

## design

설계 결정 (싱글톤)

```sql
CREATE TABLE design (
    id INTEGER PRIMARY KEY CHECK(id = 1),  -- 싱글톤 보장
    data TEXT NOT NULL,                     -- JSON
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

**JSON 포맷**:
```json
{
  "architecture": "Microservices",
  "auth_method": "JWT with 1h expiry",
  "api_style": "RESTful",
  "db_schema_users": "id, email, password_hash, role, created_at",
  "caching_strategy": "Cache-aside pattern",
  "rate_limiting": "100 req/min per user"
}
```

**필수 필드**:
- `architecture`
- `auth_method`
- `api_style`

---

## state

현재 상태 (key-value)

```sql
CREATE TABLE state (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
```

**자동 관리 항목**:

| Key | 설명 | 예시 |
|-----|------|------|
| `current_project` | 현재 프로젝트 ID | `blog-api` |
| `current_feature` | 현재 Feature ID | `3` |
| `current_task` | 현재 Task ID | `42` |
| `execution_mode` | 실행 모드 | `planning`, `execution`, `stopped` |
| `last_error` | 마지막 에러 | `Task 42 failed: timeout` |

**예시**:
```sql
INSERT INTO state VALUES ('current_project', 'blog-api');
INSERT INTO state VALUES ('execution_mode', 'execution');
```

---

## 필수 설정 체크

`clari required` 명령어로 확인:

```go
type RequiredField struct {
    Table   string   // context, tech, design
    Field   string   // JSON 필드명
    Prompt  string   // 사용자 프롬프트
    Options []string // 선택지 (선택)
}

var RequiredFields = []RequiredField{
    {"context", "project_name", "프로젝트 이름을 입력하세요", nil},
    {"context", "description", "프로젝트 설명을 입력하세요", nil},
    {"tech", "backend", "백엔드 기술을 선택하세요", []string{"go", "python", "node", "java"}},
    {"tech", "frontend", "프론트엔드 기술을 선택하세요", []string{"react", "vue", "angular", "svelte"}},
    {"tech", "database", "데이터베이스를 선택하세요", []string{"postgresql", "mysql", "sqlite", "mongodb"}},
    {"design", "architecture", "아키텍처를 선택하세요", []string{"monolithic", "microservices", "serverless"}},
    {"design", "auth_method", "인증 방식을 선택하세요", []string{"jwt", "session", "oauth2"}},
    {"design", "api_style", "API 스타일을 선택하세요", []string{"rest", "graphql", "grpc"}},
}
```

---

*Database Specification v0.0.4*
