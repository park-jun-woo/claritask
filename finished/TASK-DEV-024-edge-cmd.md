# TASK-DEV-024: Edge 커맨드

## 개요
- **파일**: `internal/cmd/edge.go`
- **유형**: 신규
- **우선순위**: High
- **Phase**: 1 (핵심 기초 구조)
- **예상 LOC**: ~200

## 목적
`clari edge` 명령어 그룹 구현

## 작업 내용

### 1. 커맨드 구조

```go
var edgeCmd = &cobra.Command{
    Use:   "edge",
    Short: "Edge (dependency) management commands",
}

// 서브커맨드
var edgeAddCmd      // clari edge add
var edgeListCmd     // clari edge list
var edgeRemoveCmd   // clari edge remove
var edgeInferCmd    // clari edge infer (향후)
```

### 2. clari edge add

의존성 추가

```bash
# Task 간 의존성 (from이 to에 의존)
clari edge add --from 3 --to 2

# Feature 간 의존성
clari edge add --feature --from 2 --to 1
```

**플래그**:
- `--from`: 의존하는 ID (필수)
- `--to`: 의존되는 ID (필수)
- `--feature`: Feature Edge 여부 (기본: Task Edge)

**응답 (Task Edge)**:
```json
{
  "success": true,
  "type": "task",
  "from_id": 3,
  "to_id": 2,
  "message": "Task edge created: auth_service depends on user_model"
}
```

**응답 (Feature Edge)**:
```json
{
  "success": true,
  "type": "feature",
  "from_id": 2,
  "to_id": 1,
  "message": "Feature edge created: 결제 depends on 로그인"
}
```

**에러 (순환 의존성)**:
```json
{
  "success": false,
  "error": "Circular dependency detected",
  "cycle": [3, 2, 1, 3]
}
```

### 3. clari edge list

의존성 목록 조회

```bash
clari edge list                  # 전체
clari edge list --feature        # Feature Edge만
clari edge list --task           # Task Edge만
clari edge list --phase 1        # 특정 Phase 내 Task Edge
```

**플래그**:
- `--feature`: Feature Edge만
- `--task`: Task Edge만
- `--phase <id>`: 특정 Phase 내

**응답**:
```json
{
  "success": true,
  "feature_edges": [
    {
      "from": {"id": 2, "name": "결제"},
      "to": {"id": 1, "name": "로그인"}
    }
  ],
  "task_edges": [
    {
      "from": {"id": 2, "title": "user_model"},
      "to": {"id": 1, "title": "user_table_sql"}
    },
    {
      "from": {"id": 3, "title": "auth_service"},
      "to": {"id": 2, "title": "user_model"}
    }
  ],
  "total_feature_edges": 1,
  "total_task_edges": 2
}
```

### 4. clari edge remove

의존성 제거

```bash
clari edge remove --from 3 --to 2
clari edge remove --feature --from 2 --to 1
```

**응답**:
```json
{
  "success": true,
  "type": "task",
  "from_id": 3,
  "to_id": 2,
  "message": "Edge removed successfully"
}
```

### 5. root.go 수정
`internal/cmd/root.go`의 `init()`에 edgeCmd 등록

## 의존성
- TASK-DEV-023 (Edge 서비스)

## 완료 기준
- [ ] 모든 서브커맨드 구현됨
- [ ] 순환 의존성 에러 처리 구현됨
- [ ] JSON 출력 형식 준수
- [ ] root.go에 등록됨
- [ ] go build 성공
