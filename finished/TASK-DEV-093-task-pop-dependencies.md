# TASK-DEV-093: task pop 의존성 기반 순서

## 개요

task pop이 의존성(task_edges)을 고려하여 실행 가능한 task만 반환

## 스펙 요구사항 (04-Task.md)

```
**동작:**
1. `status = 'pending'`인 Task 중 가장 낮은 ID 선택
2. 의존하는 Task가 모두 완료되었는지 확인
3. Task의 status를 `doing`으로 변경
```

## 현재 상태

task_edges 테이블 존재하지만 pop 시 의존성 체크 없음

## 작업 내용

### 1. GetNextExecutableTask 함수

```go
func GetNextExecutableTask(db *db.DB) (*model.Task, error) {
    // 의존성이 모두 완료된 pending task 중 가장 낮은 ID
    query := `
        SELECT t.* FROM tasks t
        WHERE t.status = 'pending'
        AND NOT EXISTS (
            SELECT 1 FROM task_edges e
            JOIN tasks dep ON e.to_task_id = dep.id
            WHERE e.from_task_id = t.id
            AND dep.status != 'done'
        )
        ORDER BY t.id ASC
        LIMIT 1
    `
    // ...
}
```

### 2. PopTask 수정

```go
func PopTask(db *db.DB) (*model.TaskPopResponse, error) {
    // 기존: 단순히 pending 중 첫 번째
    // 변경: 의존성 고려
    task, err := GetNextExecutableTask(db)
    if err != nil {
        return nil, err
    }

    if task == nil {
        // 모든 task 완료 또는 의존성 대기 중
        return &model.TaskPopResponse{Task: nil}, nil
    }

    // ... 나머지 로직
}
```

### 3. 의존 Task 정보 포함

manifest에 의존 task들의 result 포함:

```go
func GetTaskDependencies(db *db.DB, taskID int64) ([]model.Dependency, error) {
    query := `
        SELECT t.id, t.title, t.result, t.target_file
        FROM tasks t
        JOIN task_edges e ON e.to_task_id = t.id
        WHERE e.from_task_id = ?
    `
    // ...
}
```

## 응답 스키마

```json
{
  "task": {...},
  "dependencies": [
    {
      "id": "1",
      "title": "user_table_sql",
      "result": "CREATE TABLE users...",
      "file": "migrations/001_users.sql"
    }
  ],
  "manifest": {...}
}
```

## 완료 조건

- [ ] GetNextExecutableTask 함수 구현
- [ ] PopTask에서 의존성 체크
- [ ] 의존 task result를 manifest에 포함
- [ ] 테스트 작성
