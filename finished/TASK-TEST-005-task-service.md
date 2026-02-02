# TASK-TEST-005: Task 서비스 테스트

## 테스트 대상
`internal/service/task_service.go`

## 테스트 코드 위치
`test/task_service_test.go`

## 테스트 시나리오

### 1. Task CRUD 테스트
- CreateTask 성공 케이스
- CreateTask ParentID 포함 케이스
- CreateTask References 배열 포함
- GetTask 성공 케이스
- GetTask 존재하지 않는 ID 에러
- ListTasks 성공 케이스

### 2. Task 상태 전이 테스트
- StartTask pending -> doing 성공
- StartTask doing 상태에서 에러
- CompleteTask doing -> done 성공
- CompleteTask pending 상태에서 에러
- FailTask doing -> failed 성공
- FailTask pending 상태에서 에러

### 3. PopTask 테스트
- PopTask 다음 pending Task 반환
- PopTask 상태 doing으로 변경
- PopTask Manifest 포함 확인
- PopTask pending Task 없을 때 nil 반환

### 4. GetTaskStatus 테스트
- 전체 카운트 정확성
- Progress 계산 정확성

## 완료 기준
- [ ] Task CRUD 테스트
- [ ] 상태 전이 테스트
- [ ] PopTask 테스트
- [ ] GetTaskStatus 테스트
- [ ] 테스트 통과
