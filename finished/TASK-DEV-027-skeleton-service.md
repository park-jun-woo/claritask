# TASK-DEV-027: Skeleton 서비스

## 개요
- **파일**: `internal/service/skeleton_service.go`
- **유형**: 신규
- **우선순위**: High
- **Phase**: 2 (FDL 시스템)
- **예상 LOC**: ~200

## 목적
생성된 스켈레톤 파일 추적 및 관리

## 작업 내용

### 1. Skeleton CRUD

```go
// CreateSkeleton - 스켈레톤 파일 등록
func CreateSkeleton(db *db.DB, featureID int64, filePath, layer string) (int64, error)

// GetSkeleton - 스켈레톤 조회
func GetSkeleton(db *db.DB, id int64) (*model.Skeleton, error)

// ListSkeletonsByFeature - Feature별 스켈레톤 목록
func ListSkeletonsByFeature(db *db.DB, featureID int64) ([]model.Skeleton, error)

// DeleteSkeleton - 스켈레톤 삭제
func DeleteSkeleton(db *db.DB, id int64) error

// DeleteSkeletonsByFeature - Feature의 모든 스켈레톤 삭제
func DeleteSkeletonsByFeature(db *db.DB, featureID int64) error
```

### 2. 파일 체크섬 관리

```go
// CalculateFileChecksum - 파일 SHA256 체크섬 계산
func CalculateFileChecksum(filePath string) (string, error)

// UpdateSkeletonChecksum - 스켈레톤 체크섬 업데이트
func UpdateSkeletonChecksum(db *db.DB, id int64, checksum string) error

// HasSkeletonChanged - 스켈레톤 파일 변경 여부 확인
func HasSkeletonChanged(db *db.DB, id int64) (bool, error)
```

### 3. 스켈레톤 파일 읽기

```go
// ReadSkeletonContent - 스켈레톤 파일 내용 읽기
func ReadSkeletonContent(filePath string) (string, error)

// GetSkeletonAtLine - 특정 라인 주변의 코드 읽기
func GetSkeletonAtLine(filePath string, line, contextLines int) (string, error)
```

### 4. TODO 위치 추출

```go
// TODOLocation - TODO 위치 정보
type TODOLocation struct {
    Line     int    // 라인 번호
    Function string // 함수명
    Content  string // TODO 내용
}

// ExtractTODOLocations - 파일에서 TODO 위치 추출
func ExtractTODOLocations(filePath string) ([]TODOLocation, error)

// GetFunctionAtLine - 특정 라인이 속한 함수명 추출
func GetFunctionAtLine(filePath string, line int) (string, error)
```

### 5. Python 스켈레톤 생성기 호출

```go
// SkeletonGeneratorResult - 생성기 결과
type SkeletonGeneratorResult struct {
    GeneratedFiles []GeneratedFile `json:"generated_files"`
    Errors         []string        `json:"errors,omitempty"`
}

type GeneratedFile struct {
    Path     string `json:"path"`
    Layer    string `json:"layer"`
    Checksum string `json:"checksum"`
}

// RunSkeletonGenerator - Python 스켈레톤 생성기 실행
func RunSkeletonGenerator(fdlPath, outputDir string, force bool) (*SkeletonGeneratorResult, error)

// RunSkeletonGeneratorDryRun - 생성될 파일 목록만 반환
func RunSkeletonGeneratorDryRun(fdlPath string) ([]string, error)
```

## 의존성
- TASK-DEV-019 (모델 확장)
- TASK-DEV-020 (DB 마이그레이션)

## 완료 기준
- [ ] Skeleton CRUD 구현됨
- [ ] 파일 체크섬 관리 구현됨
- [ ] TODO 위치 추출 구현됨
- [ ] Python 생성기 호출 구현됨
- [ ] go build 성공
- [ ] 단위 테스트 작성됨
