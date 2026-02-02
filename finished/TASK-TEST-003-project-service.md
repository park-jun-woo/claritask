# TASK-TEST-003: 프로젝트 서비스 테스트

## 테스트 대상
`internal/service/project_service.go`

## 테스트 코드 위치
`test/project_service_test.go`

## 테스트 시나리오

### 1. 프로젝트 CRUD 테스트
- CreateProject 성공 케이스
- CreateProject 중복 ID 에러
- GetProject 성공 케이스
- GetProject 프로젝트 없음 에러
- UpdateProject 성공 케이스

### 2. Context CRUD 테스트
- SetContext 신규 생성
- SetContext 업데이트 (upsert)
- GetContext 성공 케이스
- GetContext 데이터 없음 에러

### 3. Tech CRUD 테스트
- SetTech 신규 생성
- SetTech 업데이트 (upsert)
- GetTech 성공 케이스
- GetTech 데이터 없음 에러

### 4. Design CRUD 테스트
- SetDesign 신규 생성
- SetDesign 업데이트 (upsert)
- GetDesign 성공 케이스
- GetDesign 데이터 없음 에러

### 5. CheckRequired 테스트
- 모든 필드 있음 - Ready=true
- 일부 필드 누락 - Ready=false, MissingRequired 반환

### 6. SetProjectFull 테스트
- 전체 데이터 설정 성공

## 완료 기준
- [ ] 프로젝트 CRUD 테스트
- [ ] Context/Tech/Design CRUD 테스트
- [ ] CheckRequired 테스트
- [ ] SetProjectFull 테스트
- [ ] 테스트 통과
