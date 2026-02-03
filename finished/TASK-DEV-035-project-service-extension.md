# TASK-DEV-035: Project 서비스 확장

## 개요
- **파일**: `internal/service/project_service.go`
- **유형**: 수정
- **우선순위**: Medium
- **Phase**: 5 (검증)
- **예상 LOC**: +100

## 목적
필수 설정 확인 시 옵션 정보 제공 및 project plan 개선

## 작업 내용

### 1. CheckRequired 응답 개선

현재:
```json
{
  "missing_required": [
    {"field": "tech.backend", "prompt": "백엔드 기술을 선택하세요"}
  ]
}
```

개선:
```json
{
  "missing_required": [
    {
      "field": "tech.backend",
      "prompt": "백엔드 기술을 선택하세요",
      "options": ["FastAPI", "Django", "Express", "Go"],
      "examples": ["FastAPI + PostgreSQL", "Django + MySQL"],
      "custom_allowed": true
    }
  ]
}
```

### 2. MissingField 구조체 확장

```go
// MissingField - 필수 필드 누락 정보
type MissingField struct {
    Field         string   `json:"field"`
    Prompt        string   `json:"prompt"`
    Options       []string `json:"options,omitempty"`
    Examples      []string `json:"examples,omitempty"`
    CustomAllowed bool     `json:"custom_allowed,omitempty"`
    DefaultValue  string   `json:"default_value,omitempty"`
}
```

### 3. CheckRequired 함수 개선

```go
func CheckRequired(database *db.DB) (*RequiredResult, error) {
    result := &RequiredResult{Ready: true}
    var missing []MissingField

    // Context 확인
    ctx, err := GetContext(database)
    if err != nil {
        missing = append(missing, MissingField{
            Field:    "context.project_name",
            Prompt:   "프로젝트 이름을 입력하세요",
            Examples: []string{"Blog Platform", "E-commerce API", "Todo App"},
        })
        missing = append(missing, MissingField{
            Field:    "context.description",
            Prompt:   "프로젝트 설명을 입력하세요",
            Examples: []string{"사용자가 블로그 글을 작성하고 공유하는 플랫폼"},
        })
    } else {
        // ... 기존 로직
    }

    // Tech 확인
    tech, err := GetTech(database)
    if err != nil {
        missing = append(missing, MissingField{
            Field:         "tech.backend",
            Prompt:        "백엔드 기술을 선택하세요",
            Options:       []string{"FastAPI", "Django", "Express", "Go", "Spring"},
            CustomAllowed: true,
        })
        missing = append(missing, MissingField{
            Field:         "tech.frontend",
            Prompt:        "프론트엔드 기술을 선택하세요",
            Options:       []string{"React", "Vue", "Angular", "Svelte", "None"},
            CustomAllowed: true,
        })
        missing = append(missing, MissingField{
            Field:         "tech.database",
            Prompt:        "데이터베이스를 선택하세요",
            Options:       []string{"PostgreSQL", "MySQL", "MongoDB", "SQLite"},
            CustomAllowed: true,
        })
    } else {
        // ... 기존 로직
    }

    // Design 확인
    design, err := GetDesign(database)
    if err != nil {
        missing = append(missing, MissingField{
            Field:         "design.architecture",
            Prompt:        "아키텍처 패턴을 선택하세요",
            Options:       []string{"Monolithic", "Microservices", "Serverless"},
            DefaultValue:  "Monolithic",
        })
        missing = append(missing, MissingField{
            Field:         "design.auth_method",
            Prompt:        "인증 방식을 선택하세요",
            Options:       []string{"JWT", "Session", "OAuth", "None"},
            DefaultValue:  "JWT",
        })
        missing = append(missing, MissingField{
            Field:         "design.api_style",
            Prompt:        "API 스타일을 선택하세요",
            Options:       []string{"RESTful", "GraphQL", "gRPC"},
            DefaultValue:  "RESTful",
        })
    } else {
        // ... 기존 로직
    }

    if len(missing) > 0 {
        result.Ready = false
        result.MissingRequired = missing
    }

    return result, nil
}
```

### 4. 추천 설정 제공

```go
// GetRecommendedSettings - 프로젝트 유형별 추천 설정
type RecommendedSetting struct {
    Type        string                 `json:"type"`        // web-api, fullstack, cli
    Description string                 `json:"description"`
    Tech        map[string]string      `json:"tech"`
    Design      map[string]string      `json:"design"`
}

func GetRecommendedSettings() []RecommendedSetting {
    return []RecommendedSetting{
        {
            Type:        "web-api",
            Description: "REST API 서버",
            Tech: map[string]string{
                "backend":  "FastAPI",
                "frontend": "None",
                "database": "PostgreSQL",
            },
            Design: map[string]string{
                "architecture": "Monolithic",
                "auth_method":  "JWT",
                "api_style":    "RESTful",
            },
        },
        {
            Type:        "fullstack",
            Description: "풀스택 웹 애플리케이션",
            Tech: map[string]string{
                "backend":  "FastAPI",
                "frontend": "React",
                "database": "PostgreSQL",
            },
            Design: map[string]string{
                "architecture": "Monolithic",
                "auth_method":  "JWT",
                "api_style":    "RESTful",
            },
        },
        {
            Type:        "microservices",
            Description: "마이크로서비스 아키텍처",
            Tech: map[string]string{
                "backend":  "FastAPI",
                "frontend": "React",
                "database": "PostgreSQL",
            },
            Design: map[string]string{
                "architecture": "Microservices",
                "auth_method":  "OAuth",
                "api_style":    "gRPC",
            },
        },
    }
}
```

### 5. project plan 응답 개선

```go
func runProjectPlan(cmd *cobra.Command, args []string) error {
    // ... 기존 로직

    // 추천 설정 포함
    recommendations := service.GetRecommendedSettings()

    outputJSON(map[string]interface{}{
        "success":         true,
        "ready":           result.Ready,
        "missing_required": result.MissingRequired,
        "recommendations": recommendations,
        "next_steps": []string{
            "1. Configure required settings: clari project set '{...}'",
            "2. Or set individually: clari context set, clari tech set, clari design set",
            "3. Check configuration: clari required",
            "4. Start planning: clari project plan",
        },
    })
}
```

## 의존성
- 없음 (기존 코드 개선)

## 완료 기준
- [ ] MissingField에 options, examples 추가됨
- [ ] CheckRequired에 상세 정보 포함됨
- [ ] 추천 설정 기능 구현됨
- [ ] project plan 응답 개선됨
- [ ] go build 성공
- [ ] 기존 테스트 통과
