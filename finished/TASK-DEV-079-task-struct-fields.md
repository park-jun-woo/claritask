# TASK-DEV-079: Task 구조체 필드 추가

## 개요

specs/CLI/04-Task.md에 정의된 Task 필드들을 구현

## 현재 상태

```go
type Task struct {
    ID             string    // ← INTEGER여야 함
    FeatureID      int64
    SkeletonID     *int64
    Status         string
    Title          string
    Content        string
    TargetFile     string
    TargetLine     *int
    TargetFunction string
    Result         string
    Error          string
    // ... timestamps
}
```

## 스펙 요구사항

```json
{
  "id": "3",
  "feature_id": "1",
  "parent_id": null,        // 누락
  "status": "doing",
  "title": "auth_service",
  "level": "leaf",          // 누락
  "skill": "python",        // 누락
  "references": ["..."],    // 누락
  "content": "..."
}
```

## 작업 내용

### 1. DB 스키마 수정 (db.go)

```sql
ALTER TABLE tasks ADD COLUMN parent_id INTEGER REFERENCES tasks(id);
ALTER TABLE tasks ADD COLUMN level TEXT DEFAULT 'leaf' CHECK(level IN ('leaf', 'parent'));
ALTER TABLE tasks ADD COLUMN skill TEXT DEFAULT '';
ALTER TABLE tasks ADD COLUMN refs TEXT DEFAULT '';  -- JSON array
```

### 2. Model 수정 (models.go)

```go
type Task struct {
    ID             int64     // string → int64
    FeatureID      int64
    ParentID       *int64    // 추가
    SkeletonID     *int64
    Status         string
    Title          string
    Level          string    // 추가: leaf, parent
    Skill          string    // 추가: sql, python, go, etc
    References     []string  // 추가: 참조 파일 경로들
    Content        string
    TargetFile     string
    TargetLine     *int
    TargetFunction string
    Result         string
    Error          string
    CreatedAt      time.Time
    StartedAt      *time.Time
    CompletedAt    *time.Time
    FailedAt       *time.Time
}
```

### 3. Service 수정 (task_service.go)

- CreateTask: 새 필드 처리
- GetTask: 새 필드 조회
- PopTask: 응답에 새 필드 포함
- scanTask: refs JSON 파싱

### 4. Cmd 수정 (task.go)

- taskPushInput에 level, skill, references 추가
- 출력 JSON에 새 필드 포함

## 완료 조건

- [ ] DB 스키마에 parent_id, level, skill, refs 컬럼 추가
- [ ] Task 구조체 필드 추가
- [ ] ID 타입을 int64로 변경
- [ ] task push에서 새 필드 입력 가능
- [ ] task pop/get에서 새 필드 반환
- [ ] 테스트 통과
