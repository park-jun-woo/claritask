# TASK-DEV-019: 모델 확장

## 개요
- **파일**: `internal/model/models.go`
- **유형**: 수정
- **우선순위**: High
- **Phase**: 1 (핵심 기초 구조)
- **예상 LOC**: +150

## 목적
Feature, FeatureEdge, TaskEdge, Skeleton 모델 및 Task Pop Response 확장 모델 추가

## 작업 내용

### 1. Feature 모델 추가
```go
// Feature - Feature 테이블
type Feature struct {
    ID                int64
    ProjectID         string
    Name              string
    Description       string
    Spec              string // Feature 상세 스펙
    FDL               string // FDL YAML 원문
    FDLHash           string // FDL 변경 감지용 해시
    SkeletonGenerated bool   // 스켈레톤 생성 완료 여부
    Status            string // pending, active, done
    CreatedAt         time.Time
}
```

### 2. FeatureEdge 모델 추가
```go
// FeatureEdge - Feature 간 의존성
type FeatureEdge struct {
    FromFeatureID int64
    ToFeatureID   int64
    CreatedAt     time.Time
}
```

### 3. TaskEdge 모델 추가
```go
// TaskEdge - Task 간 의존성
type TaskEdge struct {
    FromTaskID int64
    ToTaskID   int64
    CreatedAt  time.Time
}
```

### 4. Skeleton 모델 추가
```go
// Skeleton - 생성된 스켈레톤 파일 추적
type Skeleton struct {
    ID        int64
    FeatureID int64
    FilePath  string // 생성된 파일 경로
    Layer     string // model, service, api, ui
    Checksum  string // 파일 변경 감지용
    CreatedAt time.Time
}
```

### 5. Task 모델 확장
기존 Task 구조체에 필드 추가:
```go
type Task struct {
    // ... 기존 필드들
    FeatureID      *int64  // Feature ID (nullable)
    SkeletonID     *int64  // Skeleton ID (nullable)
    TargetFile     string  // 구현 대상 파일 경로
    TargetLine     *int    // 구현 대상 라인 번호
    TargetFunction string  // 구현 대상 함수명
}
```

### 6. TaskPopResponse 확장
```go
// FDLInfo - FDL 정보 (task pop 시 반환)
type FDLInfo struct {
    Feature string                 `json:"feature"`
    Service map[string]interface{} `json:"service,omitempty"`
    Models  map[string]interface{} `json:"models,omitempty"`
    API     map[string]interface{} `json:"api,omitempty"`
    UI      map[string]interface{} `json:"ui,omitempty"`
}

// SkeletonInfo - 스켈레톤 정보 (task pop 시 반환)
type SkeletonInfo struct {
    File    string `json:"file"`
    Line    int    `json:"line"`
    Content string `json:"current_content"`
}

// Dependency - 의존 Task 정보
type Dependency struct {
    ID     int64  `json:"id"`
    Title  string `json:"title"`
    Result string `json:"result"`
    File   string `json:"file,omitempty"`
}

// TaskPopResponse 확장
type TaskPopResponse struct {
    Task         *Task         `json:"task"`
    FDL          *FDLInfo      `json:"fdl,omitempty"`
    Skeleton     *SkeletonInfo `json:"skeleton,omitempty"`
    Dependencies []Dependency  `json:"dependencies,omitempty"`
    Manifest     Manifest      `json:"manifest"`
}
```

## 의존성
- 없음 (기초 작업)

## 완료 기준
- [ ] 모든 신규 모델이 정의됨
- [ ] Task 모델이 확장됨
- [ ] TaskPopResponse가 확장됨
- [ ] go build 성공
