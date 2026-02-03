# TASK-DEV-025: FDL 서비스

## 개요
- **파일**: `internal/service/fdl_service.go`
- **유형**: 신규
- **우선순위**: High
- **Phase**: 2 (FDL 시스템)
- **예상 LOC**: ~400

## 목적
FDL(Feature Definition Language) 파싱, 검증, 관리 로직 구현

## 의존성 추가
`go.mod`에 YAML 파서 추가:
```
gopkg.in/yaml.v3 v3.0.1
```

## 작업 내용

### 1. FDL 스키마 구조체

```go
// FDLSpec - FDL 전체 구조
type FDLSpec struct {
    Feature     string        `yaml:"feature"`
    Description string        `yaml:"description"`
    Models      []FDLModel    `yaml:"models,omitempty"`
    Service     []FDLService  `yaml:"service,omitempty"`
    API         []FDLAPI      `yaml:"api,omitempty"`
    UI          []FDLUI       `yaml:"ui,omitempty"`
}

// FDLModel - 데이터 모델 정의
type FDLModel struct {
    Name   string     `yaml:"name"`
    Table  string     `yaml:"table"`
    Fields []FDLField `yaml:"fields"`
}

type FDLField struct {
    Name        string `yaml:"name"`
    Type        string `yaml:"type"`
    Constraints string `yaml:"constraints,omitempty"`
}

// FDLService - 서비스 함수 정의
type FDLService struct {
    Name   string                 `yaml:"name"`
    Desc   string                 `yaml:"desc"`
    Input  map[string]interface{} `yaml:"input"`
    Output string                 `yaml:"output,omitempty"`
    Steps  []string               `yaml:"steps"`
}

// FDLAPI - API 엔드포인트 정의
type FDLAPI struct {
    Path     string                 `yaml:"path"`
    Method   string                 `yaml:"method"`
    Summary  string                 `yaml:"summary,omitempty"`
    Use      string                 `yaml:"use"` // service.FunctionName
    Request  map[string]interface{} `yaml:"request,omitempty"`
    Response map[string]interface{} `yaml:"response"`
}

// FDLUI - UI 컴포넌트 정의
type FDLUI struct {
    Component string                 `yaml:"component"`
    Type      string                 `yaml:"type"` // Page, Organism, Molecule
    Props     map[string]interface{} `yaml:"props,omitempty"`
    State     []string               `yaml:"state,omitempty"`
    Init      []string               `yaml:"init,omitempty"`
    View      []map[string]interface{} `yaml:"view,omitempty"`
    Parent    string                 `yaml:"parent,omitempty"`
}
```

### 2. FDL 파싱

```go
// ParseFDL - YAML 문자열을 FDLSpec으로 파싱
func ParseFDL(yamlStr string) (*FDLSpec, error)

// ParseFDLFile - 파일에서 FDL 파싱
func ParseFDLFile(filePath string) (*FDLSpec, error)
```

### 3. FDL 검증

```go
// ValidateFDL - FDL 유효성 검사
func ValidateFDL(spec *FDLSpec) error

// validateFeatureName - Feature 이름 검증
func validateFeatureName(name string) error

// validateModels - 모델 섹션 검증
func validateModels(models []FDLModel) error

// validateServices - 서비스 섹션 검증
func validateServices(services []FDLService) error

// validateAPIs - API 섹션 검증 (service.use 연결 확인)
func validateAPIs(apis []FDLAPI, services []FDLService) error

// validateUIs - UI 섹션 검증
func validateUIs(uis []FDLUI) error
```

### 4. FDL 해시 계산

```go
// CalculateFDLHash - FDL 변경 감지용 SHA256 해시 계산
func CalculateFDLHash(yamlStr string) string
```

### 5. FDL 템플릿 생성

```go
// GenerateFDLTemplate - 빈 FDL 템플릿 생성
func GenerateFDLTemplate(featureName string) string

// FDL 템플릿 상수
const fdlTemplate = `
feature: %s
description: "TODO: Feature description"

models:
  - name: TODO
    table: todos
    fields:
      - id: uuid (pk)
      - created_at: datetime (default: now)

service:
  - name: createTODO
    desc: "TODO: Service description"
    input: { }
    steps:
      - "TODO: Step 1"

api:
  - path: /api/todos
    method: POST
    use: service.createTODO
    response:
      201: { id: uuid }

ui:
  - component: TODOComponent
    type: Organism
    state:
      - items: Array
`
```

### 6. FDL → Task 매핑

```go
// FDLTaskMapping - FDL에서 Task 생성 정보 추출
type FDLTaskMapping struct {
    Title          string   // Task 제목
    Content        string   // Task 내용
    TargetFile     string   // 대상 파일 경로
    TargetFunction string   // 대상 함수명
    Layer          string   // model, service, api, ui
    Dependencies   []string // 의존 Task 힌트
}

// ExtractTaskMappings - FDL에서 Task 매핑 목록 추출
func ExtractTaskMappings(spec *FDLSpec, tech map[string]interface{}) ([]FDLTaskMapping, error)
```

## 의존성
- TASK-DEV-019 (모델 확장)
- TASK-DEV-021 (Feature 서비스)

## 완료 기준
- [ ] FDL 파싱 구현됨
- [ ] FDL 검증 구현됨
- [ ] FDL 해시 계산 구현됨
- [ ] 템플릿 생성 구현됨
- [ ] Task 매핑 추출 구현됨
- [ ] go build 성공
- [ ] 단위 테스트 작성됨
