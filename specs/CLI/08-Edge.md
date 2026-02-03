# clari edge - Edge (의존성) 관리

> **버전**: v0.0.3

## clari edge add

의존성 Edge 추가

```bash
clari edge add --from <id> --to <id>
clari edge add --feature --from <id> --to <id>
```

**플래그:**
- `--from`: 의존하는 ID (필수)
- `--to`: 의존받는 ID (필수)
- `--feature`: Feature Edge로 추가 (기본값: Task Edge)

**응답:**
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

## clari edge list

Edge 목록 조회

```bash
clari edge list
clari edge list --feature
clari edge list --task
clari edge list --feature-id <id>
```

**응답:**
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

## clari edge remove

Edge 삭제

```bash
clari edge remove --from <id> --to <id>
clari edge remove --feature --from <id> --to <id>
```

**응답:**
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

## clari edge infer

LLM 기반 Edge 추론

```bash
clari edge infer --feature <id>
clari edge infer --project
```

**플래그:**
- `--feature <id>`: Feature 내 Task Edge 추론
- `--project`: Feature 간 Edge 추론
- `--min-confidence`: 최소 확신도 (기본값: 0.7)

**응답:**
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

*Claritask Commands Reference v0.0.3 - 2026-02-03*
