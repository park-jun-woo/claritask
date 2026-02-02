# TASK-TEST-004: Phase 서비스 테스트

## 테스트 대상
`internal/service/phase_service.go`

## 테스트 코드 위치
`test/phase_service_test.go`

## 테스트 시나리오

### 1. Phase CRUD 테스트
- CreatePhase 성공 케이스
- CreatePhase 잘못된 project_id 에러
- GetPhase 성공 케이스
- GetPhase 존재하지 않는 ID 에러
- ListPhases 성공 케이스 (order_num 순서)
- ListPhases 빈 목록

### 2. Phase 상태 변경 테스트
- UpdatePhaseStatus 성공 케이스
- StartPhase pending -> active 성공
- StartPhase active 상태에서 에러
- CompletePhase active -> done 성공
- CompletePhase pending 상태에서 에러

### 3. Task 카운트 테스트
- ListPhases에서 tasks_total, tasks_done 정확성

## 완료 기준
- [ ] Phase CRUD 테스트
- [ ] 상태 전이 테스트
- [ ] Task 카운트 테스트
- [ ] 테스트 통과
