# TASK-DEV-037: FDL Diff 명령어 구현

## 개요
- **파일**: `internal/cmd/fdl.go`, `internal/service/fdl_service.go`
- **유형**: 신규
- **스펙 참조**: Claritask.md "검증" 섹션

## 배경
Claritask.md 스펙에 정의된 `clari fdl diff <feature_id>` 명령어가 미구현 상태.
FDL과 실제 코드 차이점을 출력하는 기능이 필요함.

## 구현 내용

### 1. fdl_service.go에 diff 함수 추가
```go
// DiffFDLImplementation shows differences between FDL and actual code
func DiffFDLImplementation(database *db.DB, featureID int64) (*DiffResult, error)

type DiffResult struct {
    FeatureID    int64
    FeatureName  string
    Differences  []FileDiff
    TotalChanges int
}

type FileDiff struct {
    FilePath     string
    Layer        string  // model, service, api, ui
    Changes      []Change
}

type Change struct {
    Type     string  // "added", "removed", "modified"
    Line     int
    Expected string
    Actual   string
}
```

### 2. fdl.go에 명령어 추가
```go
var fdlDiffCmd = &cobra.Command{
    Use:   "diff <feature_id>",
    Short: "Show differences between FDL and actual code",
    Args:  cobra.ExactArgs(1),
    RunE:  runFDLDiff,
}
```

### 3. 출력 형식
```json
{
  "success": true,
  "feature_id": 1,
  "differences": [
    {
      "file": "services/comment_service.py",
      "layer": "service",
      "changes": [
        {
          "type": "modified",
          "line": 15,
          "expected": "def createComment(userId: UUID, postId: UUID, content: str)",
          "actual": "def createComment(user_id: str, post_id: str, content: str)"
        }
      ]
    }
  ]
}
```

## 완료 기준
- [ ] DiffFDLImplementation 함수 구현
- [ ] clari fdl diff 명령어 동작
- [ ] 테스트 케이스 작성
