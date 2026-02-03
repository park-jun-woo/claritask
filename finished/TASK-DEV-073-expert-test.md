# TASK-DEV-073: Expert 테스트 작성

## 목표
`test/expert_service_test.go` 신규 파일 생성

## 작업 내용

### 1. 테스트 함수 목록
```go
func TestAddExpert(t *testing.T)
func TestListExperts(t *testing.T)
func TestGetExpert(t *testing.T)
func TestRemoveExpert(t *testing.T)
func TestAssignExpert(t *testing.T)
func TestUnassignExpert(t *testing.T)
func TestGetAssignedExperts(t *testing.T)
func TestExpertInManifest(t *testing.T)
```

### 2. TestAddExpert
- 정상 생성 확인
- 폴더 생성 확인
- EXPERT.md 파일 생성 확인
- 중복 ID 에러 확인
- 잘못된 ID 포맷 에러 확인

### 3. TestListExperts
- 전체 목록 조회
- assigned 필터
- available 필터

### 4. TestGetExpert
- 정상 조회
- 존재하지 않는 Expert 에러

### 5. TestRemoveExpert
- 정상 삭제
- 할당된 Expert 삭제 시도 (force=false) 에러
- 할당된 Expert 강제 삭제 (force=true)

### 6. TestAssignExpert
- 정상 할당
- 중복 할당 에러

### 7. TestUnassignExpert
- 정상 해제
- 할당되지 않은 Expert 해제 에러

### 8. TestGetAssignedExperts
- 할당된 Expert 목록 조회
- Content 포함 확인

### 9. TestExpertInManifest
- PopTaskFull 호출 시 Expert 포함 확인

## 완료 조건
- [ ] 모든 테스트 함수 작성
- [ ] go test 통과
- [ ] 커버리지 확인
