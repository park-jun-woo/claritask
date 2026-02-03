# clari context/tech/design/required - 설정 관리

> **버전**: v0.0.3

## clari context set / get

프로젝트 컨텍스트 관리

```bash
clari context set '<json>'
clari context get
```

**필수 필드:** `project_name`, `description`

**JSON 포맷:**
```json
{
  "project_name": "Blog Platform",
  "description": "Developer blogging platform",
  "target_users": "Tech bloggers",
  "deadline": "2026-03-01"
}
```

---

## clari tech set / get

기술 스택 관리

```bash
clari tech set '<json>'
clari tech get
```

**필수 필드:** `backend`, `frontend`, `database`

**JSON 포맷:**
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

## clari design set / get

설계 결정 관리

```bash
clari design set '<json>'
clari design get
```

**필수 필드:** `architecture`, `auth_method`, `api_style`

**JSON 포맷:**
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

## clari required

필수 설정 확인

```bash
clari required
```

**응답 (Ready):**
```json
{
  "success": true,
  "ready": true,
  "message": "All required fields configured"
}
```

**응답 (Not Ready):**
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

*Claritask Commands Reference v0.0.3 - 2026-02-03*
