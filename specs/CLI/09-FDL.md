# clari fdl - FDL 관리

> **버전**: v0.0.3

## clari fdl create

FDL 템플릿 파일 생성

```bash
clari fdl create <name>
```

**응답:**
```json
{
  "success": true,
  "file": "features/user_auth.fdl.yaml",
  "message": "FDL template created. Edit the file and run 'clari fdl register'"
}
```

---

## clari fdl register

FDL 파일을 Feature로 등록

```bash
clari fdl register <file>
```

**응답:**
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

## clari fdl validate

FDL 유효성 검증

```bash
clari fdl validate <feature_id>
```

**응답:**
```json
{
  "success": true,
  "feature_id": 1,
  "valid": true,
  "message": "FDL is valid"
}
```

---

## clari fdl show

FDL 내용 조회

```bash
clari fdl show <feature_id>
```

**응답:**
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

## clari fdl skeleton

FDL에서 스켈레톤 코드 생성

```bash
clari fdl skeleton <feature_id>
clari fdl skeleton <feature_id> --dry-run
clari fdl skeleton <feature_id> --force
```

**응답:**
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

## clari fdl tasks

FDL에서 Task 자동 생성

```bash
clari fdl tasks <feature_id>
```

**응답:**
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

## clari fdl verify

구현이 FDL과 일치하는지 검증

```bash
clari fdl verify <feature_id>
```

**응답:**
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

## clari fdl diff

FDL과 실제 코드 차이점 표시

```bash
clari fdl diff <feature_id>
```

**응답:**
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

*Claritask Commands Reference v0.0.3 - 2026-02-03*
