# TASK-DEV-104: Feature MD 파일 생성 기능 제거

## 목표
Feature 추가 시 md 파일 생성 로직 제거

## 변경 파일
- `cli/internal/service/feature_service.go`
- `cli/internal/cmd/feature.go`
- `cli/internal/model/models.go`

## 작업 내용

### 1. feature_service.go 수정
- `CreateFeature` 함수에서 md 파일 생성 로직 제거
- `featureMarkdownTemplate` 상수 제거
- 반환 타입을 `*CreateFeatureResult` → `int64`로 변경 (단순화)

### 2. feature.go 수정
- `runFeatureAdd`에서 `file_path` 응답 필드 제거
- `runFeatureCreate` 제거 또는 단순화

### 3. models.go 수정
- Feature 구조체에서 `FilePath`, `Content`, `ContentHash` 필드 제거 (선택적)

### 4. 관련 파일 수정
- `fdl.go`, `plan.go` 등에서 CreateFeature 호출 부분 수정

## 테스트
- `clari feature add` 실행 시 md 파일 생성 안됨 확인
- DB에는 정상 저장 확인
