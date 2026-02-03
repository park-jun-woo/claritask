# TASK-DEV-034: Verify 서비스

## 개요
- **파일**: `internal/service/verify_service.go`
- **유형**: 신규
- **우선순위**: Low
- **Phase**: 5 (검증)
- **예상 LOC**: ~300

## 목적
FDL과 실제 코드 일치성 검증 기능 구현

## 작업 내용

### 1. FDL 검증 결과 구조체

```go
// VerificationResult - FDL 검증 결과
type VerificationResult struct {
    FeatureID  int64              `json:"feature_id"`
    Valid      bool               `json:"valid"`
    Errors     []VerificationError `json:"errors,omitempty"`
    Warnings   []VerificationWarning `json:"warnings,omitempty"`
    Coverage   float64            `json:"coverage"` // 구현 완료율
}

type VerificationError struct {
    Type     string `json:"type"`     // missing, mismatch, signature
    Layer    string `json:"layer"`    // model, service, api, ui
    Expected string `json:"expected"`
    Actual   string `json:"actual,omitempty"`
    File     string `json:"file,omitempty"`
    Line     int    `json:"line,omitempty"`
}

type VerificationWarning struct {
    Type    string `json:"type"`
    Message string `json:"message"`
    File    string `json:"file,omitempty"`
}
```

### 2. FDL vs 코드 검증

```go
// VerifyFeature - Feature의 FDL과 실제 코드 일치성 검증
func VerifyFeature(db *db.DB, featureID int64) (*VerificationResult, error) {
    feature, err := GetFeature(db, featureID)
    if err != nil {
        return nil, err
    }

    fdl, err := ParseFDL(feature.FDL)
    if err != nil {
        return nil, err
    }

    result := &VerificationResult{
        FeatureID: featureID,
        Valid:     true,
    }

    // 각 레이어별 검증
    result.Errors = append(result.Errors, verifyModels(fdl.Models)...)
    result.Errors = append(result.Errors, verifyServices(fdl.Service)...)
    result.Errors = append(result.Errors, verifyAPIs(fdl.API)...)
    result.Errors = append(result.Errors, verifyUIs(fdl.UI)...)

    if len(result.Errors) > 0 {
        result.Valid = false
    }

    // 커버리지 계산
    result.Coverage = calculateCoverage(db, featureID)

    return result, nil
}
```

### 3. 모델 검증

```go
// verifyModels - FDL 모델과 실제 코드 비교
func verifyModels(models []FDLModel) []VerificationError {
    var errors []VerificationError

    for _, model := range models {
        // 1. 파일 존재 확인
        filePath := fmt.Sprintf("models/%s.py", strings.ToLower(model.Name))
        if !fileExists(filePath) {
            errors = append(errors, VerificationError{
                Type:     "missing",
                Layer:    "model",
                Expected: filePath,
            })
            continue
        }

        // 2. 클래스 정의 확인
        content, _ := ReadFile(filePath)
        if !strings.Contains(content, fmt.Sprintf("class %s", model.Name)) {
            errors = append(errors, VerificationError{
                Type:     "missing",
                Layer:    "model",
                Expected: fmt.Sprintf("class %s", model.Name),
                File:     filePath,
            })
        }

        // 3. 필드 확인
        for _, field := range model.Fields {
            if !strings.Contains(content, field.Name) {
                errors = append(errors, VerificationError{
                    Type:     "missing",
                    Layer:    "model",
                    Expected: fmt.Sprintf("field: %s", field.Name),
                    File:     filePath,
                })
            }
        }
    }

    return errors
}
```

### 4. 서비스 함수 검증

```go
// verifyServices - FDL 서비스와 실제 코드 비교
func verifyServices(services []FDLService) []VerificationError {
    var errors []VerificationError

    for _, svc := range services {
        // 1. 함수 존재 확인
        filePath := findServiceFile(svc.Name)
        if filePath == "" {
            errors = append(errors, VerificationError{
                Type:     "missing",
                Layer:    "service",
                Expected: fmt.Sprintf("function %s", svc.Name),
            })
            continue
        }

        // 2. 함수 시그니처 확인
        content, _ := ReadFile(filePath)
        expectedSig := buildExpectedSignature(svc)
        if !containsSignature(content, svc.Name) {
            errors = append(errors, VerificationError{
                Type:     "missing",
                Layer:    "service",
                Expected: expectedSig,
                File:     filePath,
            })
        }

        // 3. NotImplementedError 확인 (TODO 미구현)
        if strings.Contains(content, "NotImplementedError") {
            errors = append(errors, VerificationError{
                Type:     "not_implemented",
                Layer:    "service",
                Expected: svc.Name,
                File:     filePath,
            })
        }
    }

    return errors
}
```

### 5. API 엔드포인트 검증

```go
// verifyAPIs - FDL API와 실제 코드 비교
func verifyAPIs(apis []FDLAPI) []VerificationError {
    var errors []VerificationError

    for _, api := range apis {
        // 1. 라우터 파일 확인
        filePath := findAPIFile(api.Path)
        if filePath == "" {
            errors = append(errors, VerificationError{
                Type:     "missing",
                Layer:    "api",
                Expected: fmt.Sprintf("%s %s", api.Method, api.Path),
            })
            continue
        }

        // 2. 엔드포인트 정의 확인
        content, _ := ReadFile(filePath)
        if !containsEndpoint(content, api.Method, api.Path) {
            errors = append(errors, VerificationError{
                Type:     "missing",
                Layer:    "api",
                Expected: fmt.Sprintf("@router.%s(\"%s\")", strings.ToLower(api.Method), api.Path),
                File:     filePath,
            })
        }

        // 3. 서비스 연결 확인
        if api.Use != "" && !strings.Contains(content, extractServiceName(api.Use)) {
            errors = append(errors, VerificationError{
                Type:     "mismatch",
                Layer:    "api",
                Expected: fmt.Sprintf("calls %s", api.Use),
                File:     filePath,
            })
        }
    }

    return errors
}
```

### 6. FDL vs 코드 Diff

```go
// DiffResult - FDL과 코드의 차이점
type DiffResult struct {
    FeatureID int64       `json:"feature_id"`
    Diffs     []DiffItem  `json:"diffs"`
    Total     int         `json:"total"`
}

type DiffItem struct {
    Layer    string `json:"layer"`
    Type     string `json:"type"` // added, removed, modified
    FDL      string `json:"fdl,omitempty"`
    Code     string `json:"code,omitempty"`
    File     string `json:"file,omitempty"`
}

// DiffFeature - FDL과 실제 코드의 차이점 출력
func DiffFeature(db *db.DB, featureID int64) (*DiffResult, error) {
    // FDL에 정의되었지만 코드에 없는 것
    // 코드에 있지만 FDL에 없는 것
    // 시그니처가 다른 것
}
```

### 7. 커버리지 계산

```go
// calculateCoverage - 구현 완료율 계산
func calculateCoverage(db *db.DB, featureID int64) float64 {
    // Task 기반 계산
    tasks, _ := ListTasksByFeature(db, featureID)
    if len(tasks) == 0 {
        return 0.0
    }

    done := 0
    for _, t := range tasks {
        if t.Status == "done" {
            done++
        }
    }

    return float64(done) / float64(len(tasks)) * 100
}
```

## 의존성
- TASK-DEV-025 (FDL 서비스)
- TASK-DEV-027 (Skeleton 서비스)

## 완료 기준
- [ ] FDL vs 코드 검증 구현됨
- [ ] 모델 검증 구현됨
- [ ] 서비스 검증 구현됨
- [ ] API 검증 구현됨
- [ ] Diff 출력 구현됨
- [ ] 커버리지 계산 구현됨
- [ ] go build 성공
- [ ] 단위 테스트 작성됨
