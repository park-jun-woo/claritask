# TASK-DEV-103: Feature Create 통합 명령어

## 목표
`clari feature create` 명령어 구현 - Feature + FDL + Task 한 번에 생성

## 변경 파일
- `cli/internal/cmd/feature.go`
- `cli/internal/service/feature_service.go`

## 작업 내용

### 1. feature.go에 create 서브커맨드 추가
```go
var featureCreateCmd = &cobra.Command{
    Use:   "create '<json>'",
    Short: "Create feature with FDL and tasks",
    Args:  cobra.ExactArgs(1),
    RunE:  runFeatureCreate,
}
```

### 2. CreateFeatureInput 구조체
```go
type CreateFeatureInput struct {
    Name             string `json:"name"`
    Description      string `json:"description"`
    FDL              string `json:"fdl,omitempty"`
    GenerateTasks    bool   `json:"generate_tasks"`
    GenerateSkeleton bool   `json:"generate_skeleton"`
}
```

### 3. CreateFeatureResponse 구조체
```go
type CreateFeatureResponse struct {
    Success        bool     `json:"success"`
    FeatureID      int64    `json:"feature_id"`
    Name           string   `json:"name"`
    FilePath       string   `json:"file_path"`
    FDLHash        string   `json:"fdl_hash,omitempty"`
    FDLValid       bool     `json:"fdl_valid"`
    FDLErrors      []string `json:"fdl_errors,omitempty"`
    TasksCreated   int      `json:"tasks_created"`
    EdgesCreated   int      `json:"edges_created"`
    SkeletonFiles  []string `json:"skeleton_files,omitempty"`
    Message        string   `json:"message"`
}
```

### 4. CreateFeature 서비스 함수
```go
func (s *FeatureService) CreateFeature(input CreateFeatureInput) (*CreateFeatureResponse, error) {
    // 1. AddFeature 호출 (DB + md 파일 생성)
    // 2. FDL 있으면:
    //    - FDL 저장 및 검증
    //    - generate_tasks가 true면 fdl tasks 실행
    //    - generate_skeleton이 true면 fdl skeleton 실행
    // 3. 통합 응답 반환
}
```

## 테스트
```bash
clari feature create '{"name":"user_auth","description":"사용자 인증","fdl":"feature: user_auth\nversion: 1.0.0","generate_tasks":true}'
```

## 관련 스펙
- specs/CLI/07-Feature.md (v0.0.6)
