# TASK-DEV-038: Edge Infer 명령어 구현

## 개요
- **파일**: `internal/cmd/edge.go`, `internal/service/edge_service.go`
- **유형**: 신규
- **스펙 참조**: Claritask.md "Edge 관리" 섹션

## 배경
Claritask.md 스펙에 정의된 edge inference 명령어들이 미구현 상태:
- `clari edge infer --feature <id>`: Feature 내 Task Edge LLM 추론
- `clari edge infer --project`: Feature 간 Edge LLM 추론

## 구현 내용

### 1. edge_service.go에 추론 함수 추가
```go
// InferTaskEdges uses LLM to infer dependencies between tasks in a feature
func InferTaskEdges(database *db.DB, featureID int64) ([]InferredEdge, error)

// InferFeatureEdges uses LLM to infer dependencies between features
func InferFeatureEdges(database *db.DB) ([]InferredEdge, error)

type InferredEdge struct {
    FromID       int64
    FromName     string
    ToID         int64
    ToName       string
    Reason       string  // LLM이 추론한 의존 이유
    Confidence   float64 // 확신도 0.0 ~ 1.0
}
```

### 2. edge.go에 명령어 추가
```go
var edgeInferCmd = &cobra.Command{
    Use:   "infer",
    Short: "Infer dependency edges using LLM",
    RunE:  runEdgeInfer,
}

// Flags:
// --feature <id>: Feature 내 Task 의존성 추론
// --project: Feature 간 의존성 추론
// --auto-add: 추론된 edge 자동 추가
// --min-confidence: 최소 확신도 필터 (default: 0.7)
```

### 3. LLM 프롬프트 구성
- Task 목록과 내용을 제공
- 의존 관계 추론 요청
- JSON 형식으로 응답 파싱

### 4. 응답 형식
```json
{
  "success": true,
  "inferred_edges": [
    {
      "from_id": 2,
      "from_name": "user_model",
      "to_id": 1,
      "to_name": "user_table_sql",
      "reason": "user_model은 user_table_sql에서 생성된 테이블 스키마를 참조함",
      "confidence": 0.95
    }
  ],
  "total": 1
}
```

## 완료 기준
- [ ] InferTaskEdges 함수 구현
- [ ] InferFeatureEdges 함수 구현
- [ ] clari edge infer --feature 명령어 동작
- [ ] clari edge infer --project 명령어 동작
- [ ] 테스트 케이스 작성
