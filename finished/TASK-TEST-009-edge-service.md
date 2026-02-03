# TASK-TEST-009: Edge 서비스 테스트

## 개요
- **파일**: `test/edge_service_test.go`
- **유형**: 신규
- **대상**: `internal/service/edge_service.go`

## 테스트 케이스

### 1. Task Edge CRUD
- AddTaskEdge: 의존성 추가
- RemoveTaskEdge: 의존성 제거
- GetTaskEdges: 전체 Edge 조회

### 2. 의존성 분석
- GetTaskDependencies: 의존하는 Task 목록
- GetTaskDependents: 의존받는 Task 목록
- GetDependencyResults: 의존 Task 결과 조회

### 3. 순환 검사
- CheckTaskCycle: 순환 감지
- DetectAllCycles: 전체 그래프 순환 검사

### 4. Topological Sort
- TopologicalSortTasks: Phase 내 Task 정렬
- TopologicalSortFeatures: Feature 정렬

### 5. 실행 가능 Task
- GetExecutableTasks: 의존성 해결된 Task 목록
- IsTaskExecutable: 실행 가능 여부 확인

## 완료 기준
- [ ] 모든 테스트 케이스 작성됨
- [ ] go test 통과
