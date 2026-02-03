# TASK-DEV-036: FDL Verify 명령어 구현

## 개요
- **파일**: `internal/cmd/fdl.go`, `internal/service/fdl_service.go`
- **유형**: 신규
- **스펙 참조**: Claritask.md "검증" 섹션

## 배경
Claritask.md 스펙에 정의된 `clari fdl verify <feature_id>` 명령어가 미구현 상태.
구현이 FDL과 일치하는지 검사하는 기능이 필요함.

## 구현 내용

### 1. fdl_service.go에 검증 함수 추가
```go
// VerifyFDLImplementation checks if code matches FDL spec
func VerifyFDLImplementation(database *db.DB, featureID int64) (*VerifyResult, error)

type VerifyResult struct {
    Valid            bool
    Errors           []string
    Warnings         []string
    FunctionsMissing []string  // FDL에 있지만 코드에 없음
    FunctionsExtra   []string  // 코드에 있지만 FDL에 없음
    SignatureMismatch []SignatureDiff  // 시그니처 불일치
}

type SignatureDiff struct {
    Function     string
    Expected     string  // FDL 정의
    Actual       string  // 실제 코드
}
```

### 2. 검증 로직
- 스켈레톤 파일 존재 확인
- 함수 시그니처 일치 여부 확인
- API 경로 일치 여부 확인
- 모델 필드 일치 여부 확인

### 3. fdl.go에 명령어 추가
```go
var fdlVerifyCmd = &cobra.Command{
    Use:   "verify <feature_id>",
    Short: "Verify implementation matches FDL",
    Args:  cobra.ExactArgs(1),
    RunE:  runFDLVerify,
}
```

## 완료 기준
- [ ] VerifyFDLImplementation 함수 구현
- [ ] clari fdl verify 명령어 동작
- [ ] 테스트 케이스 작성
