# TASK-DEV-030: Orchestrator 서비스

## 개요
- **파일**: `internal/service/orchestrator_service.go`
- **유형**: 신규
- **우선순위**: High
- **Phase**: 3 (자동 실행)
- **예상 LOC**: ~400

## 목적
Claritask 드라이버 모드 - 자동 Task 실행 오케스트레이션 구현

## 작업 내용

### 1. 실행 상태 관리

```go
// ExecutionState - 실행 상태
type ExecutionState struct {
    Running       bool      `json:"running"`
    StartedAt     time.Time `json:"started_at"`
    CurrentTask   *int64    `json:"current_task"`
    TotalTasks    int       `json:"total_tasks"`
    CompletedTasks int      `json:"completed_tasks"`
    FailedTasks   int       `json:"failed_tasks"`
}

// GetExecutionState - 실행 상태 조회
func GetExecutionState(db *db.DB) (*ExecutionState, error)

// StartExecution - 실행 시작 기록
func StartExecution(db *db.DB) error

// StopExecution - 실행 중단 요청
func StopExecution(db *db.DB) error

// IsStopRequested - 중단 요청 확인
func IsStopRequested(db *db.DB) bool
```

### 2. 실행 계획 (Dry Run)

```go
// ExecutionPlan - 실행 계획
type ExecutionPlan struct {
    Tasks []PlannedTask `json:"tasks"`
    Total int           `json:"total"`
}

type PlannedTask struct {
    ID        int64   `json:"id"`
    Title     string  `json:"title"`
    Feature   string  `json:"feature,omitempty"`
    Phase     string  `json:"phase"`
    DependsOn []int64 `json:"depends_on,omitempty"`
    Order     int     `json:"order"`
}

// GenerateExecutionPlan - 실행 계획 생성 (dry-run용)
func GenerateExecutionPlan(db *db.DB, featureID *int64) (*ExecutionPlan, error) {
    // 1. Feature Edge 기반 Feature 순서 결정
    // 2. 각 Feature 내 Task Edge 기반 Task 순서 결정
    // 3. Topological Sort로 전체 순서 결정
    // 4. 실행 계획 반환
}
```

### 3. Claude 비대화형 호출

```go
// ClaudeResult - Claude 실행 결과
type ClaudeResult struct {
    Success bool   `json:"success"`
    Output  string `json:"output"`
    Error   string `json:"error,omitempty"`
}

// ExecuteTaskWithClaude - Task를 Claude에게 전달하여 실행
func ExecuteTaskWithClaude(task *model.Task, manifest *model.Manifest) (*ClaudeResult, error) {
    // 1. 프롬프트 구성
    prompt := buildTaskPrompt(task, manifest)

    // 2. claude --print 실행
    cmd := exec.Command("claude", "--print", prompt)
    output, err := cmd.CombinedOutput()

    // 3. 결과 파싱
    return parseClaudeResult(output, err)
}

// buildTaskPrompt - Task 실행을 위한 프롬프트 구성
func buildTaskPrompt(task *model.Task, manifest *model.Manifest) string {
    // FDL, Skeleton, Dependencies, Context 등을 포함한 프롬프트 생성
}
```

### 4. 전체 실행 루프

```go
// ExecuteAllTasks - 전체 Task 자동 실행
func ExecuteAllTasks(db *db.DB, options ExecutionOptions) error {
    StartExecution(db)
    defer StopExecution(db)

    for {
        // 1. 중단 요청 확인
        if IsStopRequested(db) {
            return nil
        }

        // 2. 다음 실행 가능 Task 조회
        task, manifest, err := GetNextExecutableTaskWithManifest(db)
        if err != nil {
            return err
        }
        if task == nil {
            // 모든 Task 완료
            return nil
        }

        // 3. Task 시작
        StartTask(db, task.ID)
        UpdateExecutionState(db, task.ID)

        // 4. Claude 실행
        result, err := ExecuteTaskWithClaude(task, manifest)

        // 5. 결과 처리
        if result.Success {
            CompleteTask(db, task.ID, result.Output)
        } else {
            // 실패 처리
            if options.FallbackInteractive {
                // 대화형 모드로 전환
                err = RunInteractiveDebugging(db, task)
            } else {
                FailTask(db, task.ID, result.Error)
                return fmt.Errorf("task %d failed: %s", task.ID, result.Error)
            }
        }
    }
}

// ExecutionOptions - 실행 옵션
type ExecutionOptions struct {
    FeatureID           *int64 // 특정 Feature만 실행
    DryRun              bool   // 실행 계획만 출력
    FallbackInteractive bool   // 실패 시 대화형 전환
}
```

### 5. Feature 단위 실행

```go
// ExecuteFeature - 특정 Feature의 Task만 실행
func ExecuteFeature(db *db.DB, featureID int64) error {
    // 1. Feature 의존성 확인
    deps, _ := GetFeatureDependencies(db, featureID)
    for _, dep := range deps {
        if dep.Status != "done" {
            return fmt.Errorf("feature %d blocked by feature %d (%s)", featureID, dep.ID, dep.Name)
        }
    }

    // 2. Feature 내 Task만 실행
    options := ExecutionOptions{FeatureID: &featureID}
    return ExecuteAllTasks(db, options)
}
```

### 6. 진행 상황 모니터링

```go
// ExecutionProgress - 실행 진행 상황
type ExecutionProgress struct {
    TotalFeatures     int     `json:"total_features"`
    CompletedFeatures int     `json:"completed_features"`
    TotalTasks        int     `json:"total_tasks"`
    Pending           int     `json:"pending"`
    Doing             int     `json:"doing"`
    Done              int     `json:"done"`
    Failed            int     `json:"failed"`
    Progress          float64 `json:"progress"`
    CurrentTask       *TaskInfo `json:"current_task,omitempty"`
    FailedTasks       []TaskInfo `json:"failed_tasks,omitempty"`
}

// GetExecutionProgress - 실행 진행 상황 조회
func GetExecutionProgress(db *db.DB) (*ExecutionProgress, error)
```

## 의존성
- TASK-DEV-023 (Edge 서비스) - Topological Sort
- TASK-DEV-029 (Task 서비스 확장) - PopTask with manifest

## 완료 기준
- [ ] 실행 상태 관리 구현됨
- [ ] 실행 계획 생성 구현됨
- [ ] Claude 비대화형 호출 구현됨
- [ ] 전체 실행 루프 구현됨
- [ ] Feature 단위 실행 구현됨
- [ ] 진행 상황 모니터링 구현됨
- [ ] go build 성공
- [ ] 단위 테스트 작성됨
