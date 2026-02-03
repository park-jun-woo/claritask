# TASK-DEV-070: Expert Service 구현

## 목표
`internal/service/expert_service.go` 신규 파일 생성

## 작업 내용

### 1. 상수 정의
```go
const (
    ExpertsDir     = ".claritask/experts"
    ExpertFileName = "EXPERT.md"
)
```

### 2. Expert 생성 (AddExpert)
```go
func AddExpert(database *db.DB, expertID string) (*model.Expert, error)
```
- expertID 유효성 검사 (영문 소문자, 숫자, 하이픈만)
- `.claritask/experts/<expert-id>/` 폴더 생성
- `EXPERT.md` 템플릿 파일 생성
- DB에 메타데이터 저장

### 3. Expert 목록 (ListExperts)
```go
func ListExperts(database *db.DB, filter string) ([]model.Expert, error)
```
- filter: "all", "assigned", "available"
- 파일 시스템 스캔 + DB 조회 병합

### 4. Expert 조회 (GetExpert)
```go
func GetExpert(database *db.DB, expertID string) (*model.Expert, error)
```
- DB에서 메타데이터 조회
- EXPERT.md 파일 파싱하여 정보 보완

### 5. Expert 삭제 (RemoveExpert)
```go
func RemoveExpert(database *db.DB, expertID string, force bool) error
```
- force=false: 할당된 경우 에러
- force=true: 할당 해제 후 삭제
- 폴더 전체 삭제
- DB 레코드 삭제

### 6. Expert 할당 (AssignExpert)
```go
func AssignExpert(database *db.DB, projectID, expertID string) error
```
- project_experts 테이블에 추가
- 중복 할당 시 에러

### 7. Expert 해제 (UnassignExpert)
```go
func UnassignExpert(database *db.DB, projectID, expertID string) error
```
- project_experts 테이블에서 삭제

### 8. 할당된 Expert 조회 (GetAssignedExperts)
```go
func GetAssignedExperts(database *db.DB, projectID string) ([]model.ExpertInfo, error)
```
- 프로젝트에 할당된 Expert 목록
- EXPERT.md 내용 포함

### 9. EXPERT.md 템플릿
```go
func getExpertTemplate(expertID string) string
```
- specs/Commands.md의 Expert 템플릿 형식 사용

### 10. EXPERT.md 파싱
```go
func parseExpertMetadata(filePath string) (*model.Expert, error)
```
- Metadata 테이블에서 ID, Name, Version, Domain, Language, Framework 추출

## 완료 조건
- [ ] AddExpert 구현
- [ ] ListExperts 구현
- [ ] GetExpert 구현
- [ ] RemoveExpert 구현
- [ ] AssignExpert 구현
- [ ] UnassignExpert 구현
- [ ] GetAssignedExperts 구현
- [ ] 템플릿 생성 함수 구현
- [ ] 메타데이터 파싱 함수 구현
- [ ] 컴파일 성공
