# clari memo - 메모 관리

> **버전**: v0.0.3

## clari memo set

Memo 저장 (upsert)

```bash
clari memo set <key> '<json>'
```

**Key 포맷:**

| 포맷 | Scope | 예시 |
|------|-------|------|
| `key` | project | `jwt_security` |
| `feature_id:key` | feature | `1:api_decisions` |
| `feature_id:task_id:key` | task | `1:3:blockers` |

**JSON 포맷:**
```json
{
  "value": "Use httpOnly cookies for refresh tokens",
  "priority": 1,
  "summary": "JWT 보안 정책",
  "tags": ["security", "jwt"]
}
```

**Priority:**
- `1`: 중요 - manifest에 자동 포함
- `2`: 보통 (기본값)
- `3`: 사소함

**응답:**
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

## clari memo get

Memo 조회

```bash
clari memo get <key>
```

**응답:**
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

## clari memo list

Memo 목록 조회

```bash
clari memo list           # 전체
clari memo list <scope>   # Scope별
```

**응답 (전체):**
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

## clari memo del

Memo 삭제

```bash
clari memo del <key>
```

**응답:**
```json
{
  "success": true,
  "message": "Memo deleted successfully"
}
```

---

*Claritask Commands Reference v0.0.3 - 2026-02-03*
