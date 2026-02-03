# TASK-DEV-039: Plan Features 명령어 구현

## 개요
- **파일**: `internal/cmd/project.go` 또는 신규 `internal/cmd/plan.go`
- **유형**: 신규
- **스펙 참조**: Claritask.md "Planning Phase" 섹션

## 배경
Claritask.md 워크플로우에 따르면 Planning Phase에서:
1. `clari plan features` - LLM이 Project description 기반으로 Feature 목록 산출

현재 이 명령어가 구현되어 있지 않음.

## 구현 내용

### 1. service/plan_service.go 신규 생성
```go
// PlanFeatures generates feature list from project description using LLM
func PlanFeatures(database *db.DB) (*FeaturePlan, error)

type FeaturePlan struct {
    Features     []PlannedFeature
    TotalCount   int
    Reasoning    string  // LLM의 분석 내용
}

type PlannedFeature struct {
    Name        string
    Description string
    Priority    int     // 1: 핵심, 2: 중요, 3: 부가
    Dependencies []string  // 의존하는 다른 Feature 이름
}
```

### 2. 명령어 추가
```go
var planFeaturesCmd = &cobra.Command{
    Use:   "features",
    Short: "Generate feature list from project description",
    RunE:  runPlanFeatures,
}

// 또는 project 서브커맨드로:
// clari project plan features
```

### 3. LLM 프롬프트 구성
- context, tech, design 정보 제공
- Feature 목록 생성 요청
- 우선순위와 의존관계 분석 요청

### 4. 응답 형식
```json
{
  "success": true,
  "features": [
    {
      "name": "로그인",
      "description": "사용자 인증 및 세션 관리",
      "priority": 1,
      "dependencies": []
    },
    {
      "name": "댓글",
      "description": "게시글 댓글 작성 및 조회",
      "priority": 2,
      "dependencies": ["로그인"]
    }
  ],
  "reasoning": "프로젝트 설명을 분석한 결과..."
}
```

## 완료 기준
- [ ] PlanFeatures 함수 구현
- [ ] clari plan features 명령어 동작
- [ ] 자동으로 Feature 생성 옵션 (--auto-create)
- [ ] 테스트 케이스 작성
