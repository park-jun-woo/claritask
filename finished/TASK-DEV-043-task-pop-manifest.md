# TASK-DEV-043: Task Pop Manifest 완성

## 개요
- **파일**: `internal/service/task_service.go`
- **유형**: 수정/검증
- **스펙 참조**: Claritask.md "Manifest 자동 반환" 섹션

## 배경
Claritask.md 스펙에 따르면 `clari task pop` 응답에 다음이 포함되어야 함:
1. task - Task 정보
2. fdl - 현재 Task의 FDL 정의
3. skeleton - 스켈레톤 코드 현재 상태
4. dependencies - 의존 Task들의 result + 파일 경로
5. manifest - context, tech, design, feature, memos

현재 구현 상태를 검토하고 누락된 부분 보완 필요.

## 확인/수정 내용

### 1. PopTaskFull 함수 검토
현재 반환 구조:
```go
type TaskPopResponse struct {
    Task         *Task
    Manifest     *Manifest
    FDL          *FDLInfo          // 확인 필요
    Skeleton     *SkeletonInfo     // 확인 필요
    Dependencies []Dependency      // 확인 필요
}
```

### 2. 스펙 요구사항 대비 검증
```json
{
  "task": { ... },
  "fdl": {
    "feature": "comment_system",
    "service": {
      "name": "createComment",
      "input": {...},
      "steps": [...]
    }
  },
  "skeleton": {
    "file": "services/comment_system_service.py",
    "line": 15,
    "current_content": "async def createComment..."
  },
  "dependencies": [
    {
      "id": 41,
      "title": "Comment model",
      "result": "Comment 모델 구현 완료",
      "file": "models/comment.py"
    }
  ],
  "manifest": {
    "context": {...},
    "tech": {...},
    "design": {...},
    "feature": {...},
    "memos": [...]
  }
}
```

### 3. 보완 필요 항목
- [ ] FDL 정보에서 현재 Task 관련 서비스/모델/API 정보만 추출
- [ ] Skeleton current_content 필드 (TODO 위치 주변 코드)
- [ ] Dependencies에 file 경로 포함
- [ ] Manifest에 feature 정보 포함

## 완료 기준
- [ ] PopTaskFull 응답이 스펙과 일치
- [ ] 모든 필수 필드 포함 확인
- [ ] 테스트 케이스 검증
