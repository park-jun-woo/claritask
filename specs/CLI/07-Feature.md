# clari feature - Feature 관리

> **현재 버전**: v0.0.6 ([변경이력](../HISTORY.md))

---

## Feature 파일 구조

Feature 추가 시 `features/<feature-name>.md` 파일이 자동 생성됩니다.

```
project/
├── features/
│   ├── user_auth.md
│   ├── blog_post.md
│   └── payment.md
└── .claritask/
    └── db.clt
```

### 파일 템플릿

```markdown
# <feature_name>

## 개요
<description>

## 요구사항
-

## 상세 스펙


## FDL
```yaml
# FDL 코드 작성
```

---
*Created by Claritask*
```

### 양방향 동기화

- **md 파일 → DB**: 파일 수정 시 DB의 `spec` 필드에 자동 반영
- **DB → md 파일**: CLI/VSCode에서 spec 수정 시 파일에 자동 반영
- **해시 기반 변경 감지**: 불필요한 동기화 방지

---

## clari feature list

Feature 목록 조회

```bash
clari feature list
```

**응답:**
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

## clari feature add

새 Feature 추가

```bash
clari feature add '<json>'
```

**JSON 포맷:**
```json
{
  "name": "user_auth",
  "description": "사용자 인증 시스템"
}
```

**동작:**
1. DB에 Feature 레코드 생성
2. `features/<name>.md` 파일 자동 생성 (템플릿 기반)
3. 파일 경로를 DB `file_path` 필드에 저장

**응답:**
```json
{
  "success": true,
  "feature_id": 1,
  "name": "user_auth",
  "file_path": "features/user_auth.md",
  "message": "Feature created successfully"
}
```

---

## clari feature create

Feature + FDL + Task 통합 생성 (VSCode에서 주로 사용)

```bash
clari feature create '<json>'
```

**JSON 포맷:**
```json
{
  "name": "user_auth",
  "description": "사용자 인증 시스템",
  "fdl": "feature: user_auth\nversion: 1.0.0\n...",
  "generate_tasks": true,
  "generate_skeleton": false
}
```

**동작:**
1. Feature 레코드 생성 + `features/<name>.md` 파일 생성
2. FDL 등록 및 검증
3. `generate_tasks: true`이면 FDL 기반 Task 자동 생성
4. `generate_skeleton: true`이면 스켈레톤 코드 생성

**응답:**
```json
{
  "success": true,
  "feature_id": 1,
  "name": "user_auth",
  "file_path": "features/user_auth.md",
  "fdl_hash": "abc123...",
  "fdl_valid": true,
  "tasks_created": 5,
  "edges_created": 3,
  "skeleton_files": [],
  "message": "Feature created with FDL and 5 tasks"
}
```

**에러 응답 (FDL 검증 실패 시):**
```json
{
  "success": false,
  "feature_id": 1,
  "name": "user_auth",
  "fdl_valid": false,
  "fdl_errors": ["Invalid layer structure at line 15"],
  "message": "Feature created but FDL validation failed"
}
```

**옵션:**
- `generate_tasks`: Task 자동 생성 여부 (기본: false)
- `generate_skeleton`: 스켈레톤 코드 생성 여부 (기본: false)

---

## clari feature get

Feature 상세 조회

```bash
clari feature get <id>
```

**응답:**
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

## clari feature spec

Feature 스펙 설정

```bash
clari feature spec <id> '<spec_text>'
```

**응답:**
```json
{
  "success": true,
  "feature_id": 1,
  "message": "Spec updated successfully"
}
```

---

## clari feature start

Feature 실행 시작

```bash
clari feature start <id>
```

**응답:**
```json
{
  "success": true,
  "feature_id": 1,
  "status": "active",
  "message": "Feature started"
}
```

---

## clari feature tasks

Feature의 Task 목록 조회

```bash
clari feature tasks <id>
clari feature tasks <id> --generate  # FDL 없이 LLM으로 생성
```

**응답:**
```json
{
  "success": true,
  "feature_id": 1,
  "tasks": [
    {
      "id": 1,
      "title": "Create user model",
      "status": "pending"
    }
  ],
  "total": 1
}
```

**옵션:**
- `--generate`: FDL 없이 LLM으로 Task 생성 (실험적)

**관련 명령어:**
- `clari fdl tasks <id>`: FDL 기반 Task 생성 (권장)
- `clari task list <feature_id>`: Task 목록 조회 (동일 기능)

---

*Claritask Commands Reference v0.0.6*
