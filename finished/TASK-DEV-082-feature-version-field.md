# TASK-DEV-082: Feature.version 구조체 필드 추가

## 개요

DB 스키마에는 features.version 컬럼이 있으나 Go 구조체에 누락됨

## 현재 상태

### DB 스키마 (db.go)
```sql
CREATE TABLE features (
    ...
    version INTEGER DEFAULT 1,  -- 있음
    ...
);
```

### Go 구조체 (models.go)
```go
type Feature struct {
    ID                int64
    ProjectID         string
    Name              string
    Description       string
    Spec              string
    FDL               string
    FDLHash           string
    SkeletonGenerated bool
    Status            string
    CreatedAt         time.Time
    // Version 누락!
}
```

## 작업 내용

### 1. Model 수정 (models.go)

```go
type Feature struct {
    ID                int64
    ProjectID         string
    Name              string
    Description       string
    Spec              string
    FDL               string
    FDLHash           string
    SkeletonGenerated bool
    Status            string
    Version           int       // 추가: 낙관적 잠금용
    CreatedAt         time.Time
}
```

### 2. Service 수정 (feature_service.go)

- GetFeature: version 컬럼 조회
- UpdateFeature: version 체크 및 증가
- scanFeature: version 스캔

### 3. 낙관적 잠금 구현

```go
func UpdateFeature(db *db.DB, feature *model.Feature) error {
    result, err := db.Exec(
        `UPDATE features SET name = ?, description = ?, ..., version = version + 1
         WHERE id = ? AND version = ?`,
        feature.Name, feature.Description, ..., feature.ID, feature.Version,
    )

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("concurrent modification detected")
    }
    return nil
}
```

## 완료 조건

- [ ] Feature 구조체에 Version 필드 추가
- [ ] GetFeature에서 version 조회
- [ ] UpdateFeature에서 낙관적 잠금 구현
- [ ] 동시 수정 시 에러 반환
- [ ] 테스트 작성
