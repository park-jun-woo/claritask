# TASK-DEV-029: Task 서비스 확장

## 개요
- **파일**: `internal/service/task_service.go`
- **유형**: 수정
- **우선순위**: High
- **Phase**: 2 (FDL 시스템)
- **예상 LOC**: +180

## 목적
Task Pop에 FDL, Skeleton, Dependencies 정보 추가

## 작업 내용

### 1. Task 생성 함수 확장

기존 `TaskCreateInput` 확장:

```go
type TaskCreateInput struct {
    PhaseID    int64
    FeatureID  *int64   // 추가
    SkeletonID *int64   // 추가
    ParentID   *int64
    Title      string
    Content    string
    Level      string
    Skill      string
    References []string
    TargetFile     string   // 추가
    TargetLine     *int     // 추가
    TargetFunction string   // 추가
}
```

### 2. PopTask 함수 확장

`PopTask` 함수가 반환하는 `TaskPopResponse` 확장:

```go
func PopTask(database *db.DB) (*model.TaskPopResponse, error) {
    // 1. 기존: pending Task 조회 및 시작

    // 2. 추가: FDL 정보 조회
    var fdlInfo *model.FDLInfo
    if task.FeatureID != nil {
        feature, err := GetFeature(database, *task.FeatureID)
        if err == nil && feature.FDL != "" {
            fdlInfo = parseFDLToInfo(feature.FDL)
        }
    }

    // 3. 추가: Skeleton 정보 조회
    var skeletonInfo *model.SkeletonInfo
    if task.SkeletonID != nil {
        skeleton, err := GetSkeleton(database, *task.SkeletonID)
        if err == nil {
            content, _ := ReadSkeletonAtLine(skeleton.FilePath, task.TargetLine, 10)
            skeletonInfo = &model.SkeletonInfo{
                File:    skeleton.FilePath,
                Line:    *task.TargetLine,
                Content: content,
            }
        }
    }

    // 4. 추가: Dependencies 조회
    deps, _ := GetDependencyResults(database, taskID)

    // 5. 기존 manifest 조회
    manifest := buildManifest(database)

    return &model.TaskPopResponse{
        Task:         task,
        FDL:          fdlInfo,
        Skeleton:     skeletonInfo,
        Dependencies: deps,
        Manifest:     *manifest,
    }, nil
}
```

### 3. FDL 정보 파싱 헬퍼

```go
// parseFDLToInfo - FDL YAML을 FDLInfo로 변환
func parseFDLToInfo(fdlYAML string) *model.FDLInfo {
    // YAML 파싱하여 FDLInfo 구조체로 변환
}
```

### 4. 의존성 결과 조회

```go
// GetDependencyResults - Task의 의존 Task 결과 조회
func GetDependencyResults(db *db.DB, taskID int64) ([]model.Dependency, error) {
    // task_edges 테이블에서 의존하는 Task 목록 조회
    // 각 Task의 result 필드 반환

    edges, _ := GetTaskDependencies(db, taskID)
    var deps []model.Dependency
    for _, edge := range edges {
        task, err := GetTask(db, edge.ID)
        if err == nil {
            deps = append(deps, model.Dependency{
                ID:     task.ID,
                Title:  task.Title,
                Result: task.Result,
                File:   task.TargetFile,
            })
        }
    }
    return deps, nil
}
```

### 5. 실행 가능 Task 조회 개선

```go
// GetNextExecutableTask - 의존성 해결된 다음 Task
func GetNextExecutableTask(db *db.DB) (*model.Task, error) {
    // 1. 모든 pending Task 조회
    // 2. 각 Task의 의존 Task 확인
    // 3. 의존 Task가 모두 done인 첫 번째 Task 반환
}
```

### 6. Task 목록 조회 확장

```go
// TaskListItem - 확장된 Task 목록 아이템
type TaskListItem struct {
    ID             int64   `json:"id"`
    PhaseID        int64   `json:"phase_id"`
    FeatureID      *int64  `json:"feature_id,omitempty"`
    Title          string  `json:"title"`
    Status         string  `json:"status"`
    TargetFile     string  `json:"target_file,omitempty"`
    TargetFunction string  `json:"target_function,omitempty"`
    DependsOn      []int64 `json:"depends_on,omitempty"`
}

// ListTasksWithDependencies - 의존성 포함 Task 목록
func ListTasksWithDependencies(db *db.DB, phaseID int64) ([]TaskListItem, error)
```

## 의존성
- TASK-DEV-019 (모델 확장)
- TASK-DEV-021 (Feature 서비스)
- TASK-DEV-023 (Edge 서비스)
- TASK-DEV-025 (FDL 서비스)
- TASK-DEV-027 (Skeleton 서비스)

## 완료 기준
- [ ] TaskCreateInput 확장됨
- [ ] PopTask에 FDL/Skeleton/Dependencies 추가됨
- [ ] 의존성 결과 조회 구현됨
- [ ] 실행 가능 Task 조회 개선됨
- [ ] go build 성공
- [ ] 기존 테스트 통과
- [ ] 신규 테스트 작성됨
