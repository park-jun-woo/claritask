# TASK-TEST-007: State 서비스 테스트

## 테스트 대상
`internal/service/state_service.go`

## 테스트 코드 위치
`test/state_service_test.go`

## 테스트 시나리오

### 1. State CRUD 테스트
- SetState 신규 생성
- SetState 업데이트 (upsert)
- GetState 성공 케이스
- GetState 존재하지 않는 키 빈 문자열 반환
- GetAllStates 전체 목록 반환
- DeleteState 성공 케이스
- DeleteState 존재하지 않는 키 에러

### 2. 상수 정의 테스트
- StateCurrentProject 상수 값 확인
- StateCurrentPhase 상수 값 확인
- StateCurrentTask 상수 값 확인
- StateNextTask 상수 값 확인

### 3. UpdateCurrentState 테스트
- 모든 상태 값 업데이트 확인
- nextTaskID=0 일 때 빈 문자열 설정

### 4. InitState 테스트
- 초기 상태 설정 확인

## 완료 기준
- [ ] State CRUD 테스트
- [ ] 상수 값 테스트
- [ ] UpdateCurrentState 테스트
- [ ] InitState 테스트
- [ ] 테스트 통과
