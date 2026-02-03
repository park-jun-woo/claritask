# TASK-DEV-096: FDL verify/diff 기능 완성

## 개요

fdl verify와 fdl diff 명령어가 specs대로 동작하는지 확인 및 완성

## 스펙 요구사항 (09-FDL.md)

### fdl verify
```json
{
  "success": true,
  "feature_id": 1,
  "valid": false,
  "errors": ["Found 2 verification issues"],
  "functions_missing": ["createUser"],
  "files_missing": ["models/user.py"]
}
```

### fdl diff
```json
{
  "success": true,
  "feature_id": 1,
  "feature_name": "user_auth",
  "differences": [
    {
      "file_path": "services/user_auth_service.py",
      "layer": "service",
      "changes": [
        {"type": "missing", "element": "function", "expected": "createUser"}
      ]
    }
  ],
  "total_changes": 1
}
```

## 현재 상태

fdl_service.go에 VerifyFDLImplementation, DiffFDLImplementation 함수 존재

## 작업 내용

### 1. VerifyFDLImplementation 개선

```go
type VerifyResult struct {
    Valid            bool
    Errors           []string
    FunctionsMissing []string
    FilesMissing     []string
}

func VerifyFDLImplementation(db *db.DB, featureID int64) (*VerifyResult, error) {
    result := &VerifyResult{Valid: true}

    // 1. Feature의 FDL 가져오기
    feature, err := GetFeature(db, featureID)
    if err != nil {
        return nil, err
    }

    // 2. FDL 파싱
    spec, err := ParseFDL(feature.FDL)
    if err != nil {
        return nil, err
    }

    // 3. 스켈레톤 파일 목록 가져오기
    skeletons, err := GetSkeletonsByFeature(db, featureID)
    if err != nil {
        return nil, err
    }

    // 4. 각 스켈레톤 파일 존재 확인
    for _, skeleton := range skeletons {
        if _, err := os.Stat(skeleton.FilePath); os.IsNotExist(err) {
            result.FilesMissing = append(result.FilesMissing, skeleton.FilePath)
            result.Valid = false
        }
    }

    // 5. 서비스 함수 존재 확인 (파일 파싱)
    for _, svc := range spec.Services {
        // 해당 파일에서 함수 찾기
        if !functionExists(svc.Name) {
            result.FunctionsMissing = append(result.FunctionsMissing, svc.Name)
            result.Valid = false
        }
    }

    // 6. 에러 메시지 생성
    if !result.Valid {
        total := len(result.FilesMissing) + len(result.FunctionsMissing)
        result.Errors = append(result.Errors, fmt.Sprintf("Found %d verification issues", total))
    }

    return result, nil
}
```

### 2. DiffFDLImplementation 개선

```go
type DiffResult struct {
    FeatureID   int64
    FeatureName string
    Differences []FileDiff
    TotalChanges int
}

type FileDiff struct {
    FilePath string
    Layer    string
    Changes  []Change
}

type Change struct {
    Type     string  // missing, extra, modified
    Element  string  // function, class, field
    Expected string
    Actual   string
}

func DiffFDLImplementation(db *db.DB, featureID int64) (*DiffResult, error) {
    result := &DiffResult{FeatureID: featureID}

    // 1. Feature 정보
    feature, _ := GetFeature(db, featureID)
    result.FeatureName = feature.Name

    // 2. FDL 파싱
    spec, _ := ParseFDL(feature.FDL)

    // 3. 각 레이어별 diff
    // Model diff
    for _, model := range spec.Models {
        diff := diffModel(model)
        if len(diff.Changes) > 0 {
            result.Differences = append(result.Differences, diff)
            result.TotalChanges += len(diff.Changes)
        }
    }

    // Service diff
    // API diff
    // UI diff

    return result, nil
}
```

## 완료 조건

- [ ] VerifyFDLImplementation에서 파일 존재 확인
- [ ] VerifyFDLImplementation에서 함수 존재 확인
- [ ] DiffFDLImplementation에서 변경사항 목록 반환
- [ ] 응답 스키마가 specs와 일치
- [ ] 테스트 작성
