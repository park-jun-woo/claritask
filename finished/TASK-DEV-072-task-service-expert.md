# TASK-DEV-072: Task Service Expert 연동

## 목표
`internal/service/task_service.go` 수정하여 Expert를 manifest에 포함

## 작업 내용

### 1. PopTaskFull 함수 수정
```go
func PopTaskFull(database *db.DB) (*model.TaskPopResponse, error) {
    // ... 기존 코드 ...

    // Build manifest
    manifest := model.Manifest{
        // ... 기존 필드 ...
        Experts: []model.ExpertInfo{}, // 추가
    }

    // ... 기존 코드 ...

    // Get assigned experts (추가)
    project, _ := GetProject(database)
    if project != nil {
        experts, err := GetAssignedExperts(database, project.ID)
        if err == nil {
            manifest.Experts = experts
        }
    }

    // ... 기존 코드 ...
}
```

### 2. PopTask 함수 수정 (선택적)
- 필요시 동일하게 수정

## 완료 조건
- [ ] PopTaskFull에 Expert 포함
- [ ] manifest.Experts 필드에 할당된 Expert 정보 추가
- [ ] 컴파일 성공
- [ ] 기존 테스트 통과
