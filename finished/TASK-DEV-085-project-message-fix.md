# TASK-DEV-085: Project 메시지 수정

## 개요

"phase" 대신 "feature" 사용하도록 메시지 수정

## 현재 상태 (cmd/project.go:174)

```go
"message": "Project is ready for planning. Use 'clari phase create' to add phases."
```

## 스펙 요구사항 (03-Project.md)

```json
{
  "message": "Project is ready for planning. Use 'clari feature add' to add features."
}
```

## 작업 내용

### 1. cmd/project.go 수정

```go
// runProjectPlan 함수
outputJSON(map[string]interface{}{
    "success":    true,
    "ready":      true,
    "project_id": project.ID,
    "mode":       "planning",
    "message":    "Project is ready for planning. Use 'clari feature add' to add features.",
})
```

### 2. 다른 곳에서도 "phase" 참조 확인

```bash
grep -r "phase" cli/internal/cmd/*.go
grep -r "phase" cli/internal/service/*.go
```

모든 "phase" 참조를 "feature"로 변경 (해당하는 경우)

## 완료 조건

- [ ] project plan 메시지 수정
- [ ] 다른 곳의 "phase" 참조 확인 및 수정
- [ ] 테스트 통과
