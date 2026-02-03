# TASK-DEV-021: Feature 서비스

## 개요
- **파일**: `internal/service/feature_service.go`
- **유형**: 신규
- **우선순위**: High
- **Phase**: 1 (핵심 기초 구조)
- **예상 LOC**: ~280

## 목적
Feature CRUD 및 Feature Edge 관리를 위한 서비스 레이어 구현

## 작업 내용

### 1. Feature CRUD 함수

```go
// CreateFeature - Feature 생성
func CreateFeature(db *db.DB, projectID, name, description string) (int64, error)

// GetFeature - Feature 조회
func GetFeature(db *db.DB, id int64) (*model.Feature, error)

// ListFeatures - 프로젝트별 Feature 목록 조회
func ListFeatures(db *db.DB, projectID string) ([]model.Feature, error)

// UpdateFeature - Feature 업데이트
func UpdateFeature(db *db.DB, feature *model.Feature) error

// DeleteFeature - Feature 삭제
func DeleteFeature(db *db.DB, id int64) error
```

### 2. Feature 상태 관리

```go
// StartFeature - Feature 시작 (pending -> active)
func StartFeature(db *db.DB, id int64) error

// CompleteFeature - Feature 완료 (active -> done)
func CompleteFeature(db *db.DB, id int64) error
```

### 3. Feature Spec 관리

```go
// SetFeatureSpec - Feature Spec 설정
func SetFeatureSpec(db *db.DB, id int64, spec string) error

// GetFeatureSpec - Feature Spec 조회
func GetFeatureSpec(db *db.DB, id int64) (string, error)
```

### 4. Feature FDL 관리

```go
// SetFeatureFDL - Feature FDL 설정
func SetFeatureFDL(db *db.DB, id int64, fdl string) error

// GetFeatureFDL - Feature FDL 조회
func GetFeatureFDL(db *db.DB, id int64) (string, error)

// CalculateFDLHash - FDL 해시 계산
func CalculateFDLHash(fdl string) string
```

### 5. Feature Edge 관리

```go
// AddFeatureEdge - Feature 간 의존성 추가
// from depends on to (to가 먼저 완료되어야 함)
func AddFeatureEdge(db *db.DB, fromID, toID int64) error

// RemoveFeatureEdge - Feature 간 의존성 제거
func RemoveFeatureEdge(db *db.DB, fromID, toID int64) error

// GetFeatureEdges - Feature의 의존성 목록 조회
func GetFeatureEdges(db *db.DB, featureID int64) ([]model.FeatureEdge, error)

// GetFeatureDependencies - Feature가 의존하는 Feature 목록
func GetFeatureDependencies(db *db.DB, featureID int64) ([]model.Feature, error)

// GetFeatureDependents - Feature에 의존하는 Feature 목록
func GetFeatureDependents(db *db.DB, featureID int64) ([]model.Feature, error)

// CheckFeatureCycle - 순환 의존성 검사
func CheckFeatureCycle(db *db.DB, fromID, toID int64) (bool, []int64, error)
```

### 6. Feature 목록 조회 (확장)

```go
// FeatureListItem - Feature 목록 아이템
type FeatureListItem struct {
    ID          int64    `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Spec        string   `json:"spec,omitempty"`
    Status      string   `json:"status"`
    TasksTotal  int      `json:"tasks_total"`
    TasksDone   int      `json:"tasks_done"`
    DependsOn   []int64  `json:"depends_on,omitempty"`
}

// ListFeaturesWithStats - 통계 포함 Feature 목록
func ListFeaturesWithStats(db *db.DB, projectID string) ([]FeatureListItem, error)
```

## 의존성
- TASK-DEV-019 (모델 확장)
- TASK-DEV-020 (DB 마이그레이션)

## 완료 기준
- [ ] 모든 함수 구현됨
- [ ] 순환 의존성 검사 로직 구현됨
- [ ] go build 성공
- [ ] 단위 테스트 작성됨
