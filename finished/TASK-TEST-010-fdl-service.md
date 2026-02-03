# TASK-TEST-010: FDL 서비스 테스트

## 개요
- **파일**: `test/fdl_service_test.go`
- **유형**: 신규
- **대상**: `internal/service/fdl_service.go`

## 테스트 케이스

### 1. FDL 파싱
- ParseFDL: 유효한 YAML 파싱
- ParseFDL: 잘못된 YAML 에러

### 2. FDL 검증
- ValidateFDL: 유효한 FDL
- validateFeatureName: 잘못된 이름 에러
- validateModels: 중복 모델 에러
- validateServices: 빈 steps 에러
- validateAPIs: 존재하지 않는 서비스 참조 에러

### 3. Task 매핑
- ExtractTaskMappings: 모델, 서비스, API, UI Task 생성
- 의존성 추론 확인

### 4. 템플릿 생성
- GenerateFDLTemplate: 유효한 템플릿 생성

## 완료 기준
- [ ] 모든 테스트 케이스 작성됨
- [ ] go test 통과
