# TASK-TEST-006: Memo 서비스 테스트

## 테스트 대상
`internal/service/memo_service.go`

## 테스트 코드 위치
`test/memo_service_test.go`

## 테스트 시나리오

### 1. Memo CRUD 테스트
- SetMemo 신규 생성
- SetMemo 업데이트 (upsert)
- SetMemo Summary, Tags 포함
- GetMemo 성공 케이스
- GetMemo 존재하지 않는 키 에러
- DeleteMemo 성공 케이스
- DeleteMemo 존재하지 않는 키 에러

### 2. Memo 목록 조회 테스트
- ListMemos 전체 목록 반환
- ListMemos scope별 분류 확인
- ListMemosByScope 특정 scope 필터링

### 3. Priority 테스트
- GetHighPriorityMemos priority=1만 반환
- 기본 priority=2 확인

### 4. ParseMemoKey 테스트
- 프로젝트 레벨: "key" -> project, "", "key"
- Phase 레벨: "PH001:key" -> phase, "PH001", "key"
- Task 레벨: "PH001:T042:notes" -> task, "PH001:T042", "notes"
- 잘못된 형식 에러

## 완료 기준
- [ ] Memo CRUD 테스트
- [ ] 목록 조회 테스트
- [ ] Priority 테스트
- [ ] ParseMemoKey 테스트
- [ ] 테스트 통과
