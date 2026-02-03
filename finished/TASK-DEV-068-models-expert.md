# TASK-DEV-068: Expert 모델 추가

## 목표
`internal/model/models.go`에 Expert 관련 모델 추가

## 작업 내용

### 1. Expert 모델 추가
```go
// Expert - Expert 정의 (파일 기반, 메타데이터만 DB 저장)
type Expert struct {
    ID        string    // expert-id (폴더명)
    Name      string    // Expert 이름
    Version   string    // 버전
    Domain    string    // 도메인 설명
    Language  string    // 주 언어
    Framework string    // 주 프레임워크
    Path      string    // EXPERT.md 파일 경로
    CreatedAt time.Time
}
```

### 2. ProjectExpert 모델 추가
```go
// ProjectExpert - 프로젝트-Expert 연결
type ProjectExpert struct {
    ProjectID  string
    ExpertID   string
    AssignedAt time.Time
}
```

### 3. ExpertInfo 모델 추가 (Manifest용)
```go
// ExpertInfo - Expert 정보 (manifest 포함용)
type ExpertInfo struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Content string `json:"content"` // EXPERT.md 전체 내용
}
```

### 4. Manifest 구조체 수정
```go
type Manifest struct {
    Context map[string]interface{} `json:"context"`
    Tech    map[string]interface{} `json:"tech"`
    Design  map[string]interface{} `json:"design"`
    Feature map[string]interface{} `json:"feature,omitempty"`
    Experts []ExpertInfo           `json:"experts,omitempty"` // 추가
    State   map[string]string      `json:"state"`
    Memos   []MemoData             `json:"memos"`
}
```

## 완료 조건
- [ ] Expert 모델 정의
- [ ] ProjectExpert 모델 정의
- [ ] ExpertInfo 모델 정의
- [ ] Manifest에 Experts 필드 추가
- [ ] 컴파일 성공
