# clari feature - Feature 관리

> **버전**: v0.0.3

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

**응답:**
```json
{
  "success": true,
  "feature_id": 1,
  "name": "user_auth",
  "message": "Feature created successfully"
}
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

*Claritask Commands Reference v0.0.3 - 2026-02-03*
