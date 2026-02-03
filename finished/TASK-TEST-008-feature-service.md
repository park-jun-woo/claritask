# TASK-TEST-008: Feature 서비스 테스트

## 개요
- **파일**: `test/feature_service_test.go`
- **유형**: 신규
- **대상**: `internal/service/feature_service.go`

## 테스트 케이스

### 1. Feature CRUD
- CreateFeature: 정상 생성
- GetFeature: 존재하는 Feature 조회
- GetFeature: 존재하지 않는 Feature 조회 에러
- ListFeatures: 프로젝트별 목록 조회
- UpdateFeature: 업데이트 성공
- DeleteFeature: 삭제 성공

### 2. Feature 상태 관리
- StartFeature: pending → active 전환
- CompleteFeature: active → done 전환

### 3. FDL 해시 계산
- CalculateFDLHash: 동일 입력에 동일 해시

### 4. Feature Edge
- AddFeatureEdge: 의존성 추가
- RemoveFeatureEdge: 의존성 제거
- CheckFeatureCycle: 순환 의존성 감지

## 완료 기준
- [ ] 모든 테스트 케이스 작성됨
- [ ] go test 통과
