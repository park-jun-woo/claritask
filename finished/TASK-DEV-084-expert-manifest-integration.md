# TASK-DEV-084: Expert Manifest 연동

## 개요

task pop 시 할당된 Expert 내용을 manifest에 포함

## 스펙 요구사항 (04-Task.md)

```json
{
  "manifest": {
    "context": {...},
    "tech": {...},
    "design": {...},
    "experts": [
      {
        "id": "backend-go-gin",
        "name": "Backend Go GIN Developer",
        "content": "# Expert: Backend Go GIN Developer\n..."
      }
    ],
    "state": {...},
    "memos": [...]
  }
}
```

## 현재 상태

- Manifest 구조체에 Experts 필드 있음
- PopTask에서 experts 로딩 로직 없음

## 작업 내용

### 1. task_service.go PopTask 수정

```go
func PopTask(db *db.DB) (*model.TaskPopResponse, error) {
    // ... 기존 코드 ...

    // Manifest 구성
    manifest := model.Manifest{
        Context: contextData,
        Tech:    techData,
        Design:  designData,
        State:   states,
        Memos:   memos,
    }

    // Project 레벨 Expert 로딩
    project, _ := GetProject(db)
    if project != nil {
        experts, err := GetAssignedExperts(db, project.ID)
        if err == nil {
            manifest.Experts = experts
        }
    }

    // Feature 레벨 Expert 로딩 (있다면)
    if task != nil && task.FeatureID > 0 {
        featureExperts, err := GetAssignedExpertsByFeature(db, task.FeatureID)
        if err == nil {
            // 중복 제거 후 추가
            manifest.Experts = mergeExperts(manifest.Experts, featureExperts)
        }
    }

    return &model.TaskPopResponse{
        Task:     task,
        Manifest: manifest,
    }, nil
}
```

### 2. expert_service.go 수정

GetAssignedExperts 함수가 ExpertInfo를 반환하도록:

```go
type ExpertInfo struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Content string `json:"content"`
}

func GetAssignedExperts(db *db.DB, projectID string) ([]model.ExpertInfo, error) {
    // project_experts 테이블에서 조회
    // EXPERT.md 파일 내용 읽기
    // ExpertInfo 반환
}
```

### 3. 중복 Expert 처리

```go
func mergeExperts(a, b []model.ExpertInfo) []model.ExpertInfo {
    seen := make(map[string]bool)
    result := []model.ExpertInfo{}

    for _, e := range a {
        if !seen[e.ID] {
            seen[e.ID] = true
            result = append(result, e)
        }
    }
    for _, e := range b {
        if !seen[e.ID] {
            seen[e.ID] = true
            result = append(result, e)
        }
    }
    return result
}
```

## 완료 조건

- [ ] PopTask에서 project 레벨 expert 로딩
- [ ] PopTask에서 feature 레벨 expert 로딩 (TASK-DEV-080 후)
- [ ] manifest.experts에 content 포함
- [ ] 중복 expert 제거
- [ ] 테스트 작성
