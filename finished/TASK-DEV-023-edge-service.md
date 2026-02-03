# TASK-DEV-023: Edge 서비스

## 개요
- **파일**: `internal/service/edge_service.go`
- **유형**: 신규
- **우선순위**: High
- **Phase**: 1 (핵심 기초 구조)
- **예상 LOC**: ~250

## 목적
Task Edge CRUD 및 의존성 분석 로직 구현

## 작업 내용

### 1. Task Edge CRUD

```go
// AddTaskEdge - Task 간 의존성 추가
// from depends on to (to가 먼저 완료되어야 함)
func AddTaskEdge(db *db.DB, fromID, toID int64) error

// RemoveTaskEdge - Task 간 의존성 제거
func RemoveTaskEdge(db *db.DB, fromID, toID int64) error

// GetTaskEdges - 전체 Task Edge 목록
func GetTaskEdges(db *db.DB) ([]model.TaskEdge, error)

// GetTaskEdgesByPhase - Phase 내 Task Edge 목록
func GetTaskEdgesByPhase(db *db.DB, phaseID int64) ([]model.TaskEdge, error)

// GetTaskEdgesByFeature - Feature 내 Task Edge 목록
func GetTaskEdgesByFeature(db *db.DB, featureID int64) ([]model.TaskEdge, error)
```

### 2. 의존성 분석

```go
// GetTaskDependencies - Task가 의존하는 Task 목록
func GetTaskDependencies(db *db.DB, taskID int64) ([]model.Task, error)

// GetTaskDependents - Task에 의존하는 Task 목록
func GetTaskDependents(db *db.DB, taskID int64) ([]model.Task, error)

// GetDependencyResults - 의존 Task들의 result 조회
func GetDependencyResults(db *db.DB, taskID int64) ([]model.Dependency, error)
```

### 3. 순환 의존성 검사

```go
// CheckTaskCycle - 순환 의존성 검사
// returns: (hasCycle, cyclePath, error)
func CheckTaskCycle(db *db.DB, fromID, toID int64) (bool, []int64, error)

// DetectAllCycles - 전체 그래프에서 순환 검사
func DetectAllCycles(db *db.DB) ([][]int64, error)
```

### 4. Topological Sort

```go
// TopologicalSortTasks - Task 실행 순서 결정
// 의존성이 해결된 순서대로 정렬
func TopologicalSortTasks(db *db.DB, phaseID int64) ([]model.Task, error)

// TopologicalSortFeatures - Feature 실행 순서 결정
func TopologicalSortFeatures(db *db.DB, projectID string) ([]model.Feature, error)
```

### 5. 실행 가능 Task 조회

```go
// GetExecutableTasks - 의존성이 모두 완료된 pending Task 목록
func GetExecutableTasks(db *db.DB) ([]model.Task, error)

// IsTaskExecutable - 특정 Task가 실행 가능한지 확인
func IsTaskExecutable(db *db.DB, taskID int64) (bool, []model.Task, error)
// returns: (isExecutable, blockingTasks, error)
```

### 6. Edge 목록 조회 (통합)

```go
// EdgeListResult - Edge 목록 응답
type EdgeListResult struct {
    FeatureEdges []FeatureEdgeItem `json:"feature_edges"`
    TaskEdges    []TaskEdgeItem    `json:"task_edges"`
    TotalFeature int               `json:"total_feature_edges"`
    TotalTask    int               `json:"total_task_edges"`
}

type FeatureEdgeItem struct {
    From FeatureRef `json:"from"`
    To   FeatureRef `json:"to"`
}

type TaskEdgeItem struct {
    From TaskRef `json:"from"`
    To   TaskRef `json:"to"`
}

type FeatureRef struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

type TaskRef struct {
    ID    int64  `json:"id"`
    Title string `json:"title"`
}

// ListAllEdges - 전체 Edge 목록
func ListAllEdges(db *db.DB) (*EdgeListResult, error)
```

## 의존성
- TASK-DEV-019 (모델 확장)
- TASK-DEV-020 (DB 마이그레이션)
- TASK-DEV-021 (Feature 서비스) - FeatureEdge 관련

## 완료 기준
- [ ] 모든 함수 구현됨
- [ ] 순환 의존성 검사 구현됨 (DFS 기반)
- [ ] Topological Sort 구현됨 (Kahn's algorithm)
- [ ] go build 성공
- [ ] 단위 테스트 작성됨
