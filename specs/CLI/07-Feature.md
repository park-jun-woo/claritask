# clari feature - Feature 관리

> **현재 버전**: v0.0.7 ([변경이력](../HISTORY.md))

---

## Feature 파일 구조

Feature 추가 시 `features/<feature-name>.fdl.yaml` 파일이 생성됩니다.

```
project/
├── features/
│   ├── user_auth.fdl.yaml
│   ├── blog_post.fdl.yaml
│   └── payment.fdl.yaml
└── .claritask/
    └── db.clt
```

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
      "description": "사용자 인증 시스템",
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

새 Feature 추가 (Claude Code로 FDL 생성)

```bash
clari feature add '<json>'
```

**JSON 포맷:**
```json
{
  "name": "user_auth",
  "description": "사용자 인증 시스템. JWT 기반 로그인/로그아웃, 회원가입 기능 포함."
}
```

**동작:**
1. DB에 Feature 레코드 생성
2. **Claude Code 호출하여 FDL 생성** (TTY Handover)
3. 생성된 FDL을 `features/<name>.fdl.yaml`에 저장
4. DB에 FDL 내용 및 해시 저장

**응답 (Phase 1 - Claude Code 호출 전):**
```json
{
  "success": true,
  "feature_id": 1,
  "name": "user_auth",
  "mode": "tty_handover",
  "prompt": "다음 Feature에 대한 FDL YAML을 생성해주세요:\n\nFeature: user_auth\nDescription: 사용자 인증 시스템...",
  "message": "Feature created. Launching Claude Code for FDL generation..."
}
```

**Claude Code가 생성하는 FDL 예시:**
```yaml
feature: user_auth
version: 1.0.0
description: 사용자 인증 시스템

layers:
  data:
    models:
      - name: User
        fields:
          - name: id
            type: int
            pk: true
          - name: email
            type: string
            unique: true
          - name: password_hash
            type: string

  logic:
    services:
      - name: AuthService
        methods:
          - name: login
            input:
              email: string
              password: string
            output: token: string
            steps:
              - db.query: SELECT * FROM users WHERE email = {email}
              - validate: password matches hash
              - return: JWT token
```

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
    "description": "사용자 인증 시스템",
    "fdl": "feature: user_auth\n...",
    "fdl_hash": "abc123...",
    "skeleton_generated": true,
    "status": "pending"
  }
}
```

---

## clari feature delete

Feature 삭제

```bash
clari feature delete <id>
```

**응답:**
```json
{
  "success": true,
  "feature_id": 1,
  "name": "user_auth",
  "message": "Feature deleted successfully"
}
```

---

## clari feature spec

Feature 설명 수정

```bash
clari feature spec <id> '<description>'
```

**응답:**
```json
{
  "success": true,
  "feature_id": 1,
  "message": "Description updated successfully"
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

---

## clari feature fdl

Feature의 FDL 재생성 (Claude Code 호출)

```bash
clari feature fdl <id>
```

**동작:**
1. Feature 정보 조회
2. Claude Code 호출하여 FDL 재생성
3. `features/<name>.fdl.yaml` 파일 업데이트
4. DB 업데이트

**응답:**
```json
{
  "success": true,
  "feature_id": 1,
  "mode": "tty_handover",
  "message": "Launching Claude Code for FDL regeneration..."
}
```

---

*Claritask Commands Reference v0.0.7*
