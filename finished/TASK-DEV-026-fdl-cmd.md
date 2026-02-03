# TASK-DEV-026: FDL 커맨드

## 개요
- **파일**: `internal/cmd/fdl.go`
- **유형**: 신규
- **우선순위**: High
- **Phase**: 2 (FDL 시스템)
- **예상 LOC**: ~350

## 목적
`clari fdl` 명령어 그룹 구현

## 작업 내용

### 1. 커맨드 구조

```go
var fdlCmd = &cobra.Command{
    Use:   "fdl",
    Short: "FDL (Feature Definition Language) management",
}

// 서브커맨드
var fdlCreateCmd    // clari fdl create <name>
var fdlRegisterCmd  // clari fdl register <file>
var fdlValidateCmd  // clari fdl validate <feature_id>
var fdlShowCmd      // clari fdl show <feature_id>
var fdlSkeletonCmd  // clari fdl skeleton <feature_id>
var fdlTasksCmd     // clari fdl tasks <feature_id>
```

### 2. clari fdl create

FDL 템플릿 생성

```bash
clari fdl create comment_system
```

**동작**:
1. `features/comment_system.fdl.yaml` 파일 생성
2. 기본 FDL 템플릿 작성

**응답**:
```json
{
  "success": true,
  "file": "features/comment_system.fdl.yaml",
  "message": "FDL template created. Edit the file and run 'clari fdl register'"
}
```

### 3. clari fdl register

FDL 파일 등록 (Feature 생성)

```bash
clari fdl register features/comment_system.fdl.yaml
```

**동작**:
1. FDL 파일 읽기
2. FDL 파싱 및 검증
3. Feature 생성 (feature 테이블에 저장)
4. FDL 해시 저장

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "feature_name": "comment_system",
  "fdl_hash": "abc123...",
  "message": "FDL registered successfully"
}
```

**에러 (검증 실패)**:
```json
{
  "success": false,
  "error": "FDL validation failed",
  "details": [
    "Missing required field: description",
    "API 'POST /comments' references unknown service 'createComment'"
  ]
}
```

### 4. clari fdl validate

FDL 검증 (저장하지 않음)

```bash
clari fdl validate 1
```

**응답 (성공)**:
```json
{
  "success": true,
  "feature_id": 1,
  "valid": true,
  "message": "FDL is valid"
}
```

**응답 (실패)**:
```json
{
  "success": true,
  "feature_id": 1,
  "valid": false,
  "errors": [
    "Service 'createComment' missing input definition"
  ]
}
```

### 5. clari fdl show

FDL 내용 조회

```bash
clari fdl show 1
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "feature_name": "comment_system",
  "fdl": "feature: comment_system\ndescription: ...",
  "fdl_hash": "abc123...",
  "skeleton_generated": false
}
```

### 6. clari fdl skeleton

스켈레톤 생성 (Python 호출)

```bash
clari fdl skeleton 1
clari fdl skeleton 1 --dry-run   # 생성될 파일 목록만
clari fdl skeleton 1 --force     # 기존 덮어쓰기
```

**플래그**:
- `--dry-run`: 실제 생성 없이 목록만 출력
- `--force`: 기존 스켈레톤 덮어쓰기

**동작**:
1. Feature의 FDL 조회
2. Python 스켈레톤 생성기 호출
3. 생성된 파일을 skeletons 테이블에 등록

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "generated_files": [
    {"path": "models/comment.py", "layer": "model"},
    {"path": "services/comment_system_service.py", "layer": "service"},
    {"path": "api/comment_system_api.py", "layer": "api"}
  ],
  "total": 3,
  "message": "Skeletons generated successfully"
}
```

### 7. clari fdl tasks

FDL 기반 Task 자동 생성

```bash
clari fdl tasks 1
```

**동작**:
1. Feature의 FDL 및 스켈레톤 조회
2. 스켈레톤의 TODO 위치에서 Task 추출
3. Task 간 Edge 자동 추론 (Model → Service → API)
4. tasks 테이블에 저장

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "tasks_created": [
    {"id": 1, "title": "Implement Comment model", "target_file": "models/comment.py"},
    {"id": 2, "title": "Implement createComment", "target_file": "services/comment_system_service.py", "target_function": "createComment"},
    {"id": 3, "title": "Implement listComments", "target_file": "services/comment_system_service.py", "target_function": "listComments"}
  ],
  "edges_created": [
    {"from": 2, "to": 1},
    {"from": 3, "to": 1}
  ],
  "total_tasks": 3,
  "total_edges": 2,
  "message": "Tasks generated from FDL"
}
```

### 8. root.go 수정
`internal/cmd/root.go`의 `init()`에 fdlCmd 등록

## 의존성
- TASK-DEV-025 (FDL 서비스)
- TASK-DEV-027 (Skeleton 서비스)
- TASK-DEV-028 (Skeleton Generator - Python)

## 완료 기준
- [ ] 모든 서브커맨드 구현됨
- [ ] FDL 파일 생성/읽기 구현됨
- [ ] Python 스켈레톤 생성기 호출 구현됨
- [ ] JSON 출력 형식 준수
- [ ] root.go에 등록됨
- [ ] go build 성공
