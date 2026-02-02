# TASK-TEST-001: 데이터 모델 테스트

## 테스트 대상
`internal/model/models.go`

## 테스트 코드 위치
`test/models_test.go`

## 테스트 시나리오

### 1. 구조체 필드 검증
- Project 구조체 필드 존재 확인
- Phase 구조체 필드 존재 확인
- Task 구조체 필드 존재 확인
- Context, Tech, Design 구조체 필드 존재 확인
- State 구조체 필드 존재 확인
- Memo 구조체 필드 존재 확인

### 2. JSON 직렬화 테스트
- Response 구조체 JSON 직렬화/역직렬화
- MemoData 구조체 JSON 직렬화/역직렬화
- TaskPopResponse 구조체 JSON 직렬화/역직렬화
- Manifest 구조체 JSON 직렬화/역직렬화

### 3. 기본값 테스트
- Task.References 빈 배열 초기화 확인
- nullable 필드 nil 처리 확인

## 완료 기준
- [ ] 모든 구조체 필드 접근 테스트
- [ ] JSON 태그 올바른 동작 확인
- [ ] 테스트 통과
