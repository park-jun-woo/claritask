# TASK-DEV-102: Feature Add 시 MD 파일 자동 생성

## 목표
`clari feature add` 실행 시 `features/<name>.md` 파일 자동 생성

## 변경 파일
- `cli/internal/service/feature_service.go`

## 작업 내용

### 1. MD 파일 템플릿 정의
```go
const featureMarkdownTemplate = `# %s

## 개요
%s

## 요구사항
-

## 상세 스펙


## FDL
` + "```yaml" + `
# FDL 코드 작성
` + "```" + `

---
*Created by Claritask*
`
```

### 2. AddFeature 함수 수정
```go
func (s *FeatureService) AddFeature(input AddFeatureInput) (*AddFeatureResponse, error) {
    // 1. DB에 Feature 레코드 생성
    // 2. features/ 디렉토리 생성 (없으면)
    // 3. features/<name>.md 파일 생성
    // 4. DB에 file_path 업데이트
    // 5. 응답에 file_path 포함
}
```

### 3. 응답 구조체 수정
```go
type AddFeatureResponse struct {
    Success   bool   `json:"success"`
    FeatureID int64  `json:"feature_id"`
    Name      string `json:"name"`
    FilePath  string `json:"file_path"`  // 추가
    Message   string `json:"message"`
}
```

## 테스트
- `clari feature add '{"name":"test_feature","description":"테스트"}'`
- `features/test_feature.md` 파일 생성 확인
- 응답에 `file_path` 포함 확인

## 관련 스펙
- specs/CLI/07-Feature.md (v0.0.6)
