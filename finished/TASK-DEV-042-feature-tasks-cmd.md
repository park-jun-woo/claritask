# TASK-DEV-042: Feature Tasks 명령어 구현

## 개요
- **파일**: `internal/cmd/feature.go`
- **유형**: 신규
- **스펙 참조**: Claritask.md "Feature 관리" 섹션

## 배경
Claritask.md 스펙에 정의된 `clari feature <id> tasks` 명령어가 미구현 상태.
현재 `clari fdl tasks`는 FDL 기반 Task 생성이지만,
`clari feature <id> tasks`는 Feature 하위 Task 목록 조회 또는 일반 Task 생성이어야 함.

## 구현 내용

### 1. feature.go에 명령어 추가
```go
var featureTasksCmd = &cobra.Command{
    Use:   "tasks <id>",
    Short: "List or generate tasks for a feature",
    Args:  cobra.ExactArgs(1),
    RunE:  runFeatureTasks,
}

// Flags:
// --generate: FDL 없이 LLM으로 Task 생성
```

### 2. 기능 구분
1. **FDL이 있는 경우**: `clari fdl tasks <id>` 호출과 동일하게 동작
2. **FDL이 없는 경우**: Feature spec 기반으로 LLM이 Task 추론

### 3. 응답 형식
```json
{
  "success": true,
  "feature_id": 1,
  "feature_name": "로그인",
  "tasks": [
    {"id": 1, "title": "user_table_sql", "status": "done"},
    {"id": 2, "title": "user_model", "status": "pending"}
  ],
  "total": 2
}
```

## 완료 기준
- [ ] featureTasksCmd 명령어 추가
- [ ] Feature 하위 Task 목록 조회 기능
- [ ] --generate 플래그로 Task 생성 기능
- [ ] 테스트 케이스 작성
