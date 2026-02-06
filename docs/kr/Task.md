# Task 시스템 설계

> Claribot의 Task 기반 분할 정복 시스템

---

## 개요

대규모 작업을 재귀적으로 분할하고, 연관 Task의 컨텍스트를 주입하여 일관성 있는 실행을 보장한다.

**핵심 아이디어**: 1회차 순회에서 Claude가 분할/계획을 판단하고, DFS로 leaf까지 순회한다.

---

## Task 구조

```go
Task {
    ID          int
    ParentID    *int      // 상위 Task
    Title       string    // 제목
    Spec        string    // 요구사항 명세서 (불변)
    Plan        string    // 계획서 (leaf만)
    Report      string    // 완료 보고서 (2회차 순회 후)
    Status      string    // todo → split/planned → done
    Error       string
    IsLeaf      bool      // true: 실행 대상, false: 분할됨
    Depth       int       // 트리 깊이 (root=0)
    CreatedAt   string
    UpdatedAt   string
}
```

### 상태 전이

```
todo ─┬─→ split (분할됨, is_leaf=false)
      │
      └─→ planned (계획됨, is_leaf=true) → done
```

| 상태 | 설명 | is_leaf |
|------|------|---------|
| todo | 등록됨, 아직 처리 안됨 | true (기본값) |
| split | 분할됨, 하위 Task 있음 | false |
| planned | 계획 완료, 실행 대기 | true |
| done | 실행 완료 | true |
| failed | 실패 | - |

---

## 1회차 순회 (재귀 분할)

### 흐름

```
Task A (todo)
├─ Claude 판단: 분할 필요
├─ clari task add로 하위 Task B, C 생성
├─ Task A → split
├─ 즉시 Task B 순회
│   └─ Claude 판단: 계획 가능 → planned
└─ 즉시 Task C 순회
    ├─ Claude 판단: 분할 필요
    ├─ Task D, E 생성
    ├─ Task C → split
    ├─ Task D 순회 → planned
    └─ Task E 순회 → planned
```

### Claude 판단 기준

**분할 조건** (하나라도 해당):
- 3개 이상의 독립적 단계
- 다중 도메인 (UI + API + DB)
- 변경 파일 5개 초과 예상

**계획 조건** (모두 만족):
- 단일 목적
- 변경 파일 5개 이하
- 독립 실행 가능

### 출력 형식

```
[SPLIT]
- Task #<id>: <title>
- Task #<id>: <title>
```

또는

```
[PLANNED]
## 구현 방향
...
```

### 깊이 제한

`MaxDepth = 5` 도달 시 강제로 계획 작성.

---

## 2회차 순회 (실행)

**대상**: `is_leaf = true AND status = 'planned'`

**순서**: 깊이 깊은 것부터 (depth DESC)

```sql
SELECT * FROM tasks
WHERE is_leaf = 1 AND status = 'planned'
ORDER BY depth DESC, id ASC
```

---

## CLI 명령어

### task add

```bash
# 기본
clari task add "제목"

# 부모 지정
clari task add "제목" --parent 1

# Spec 포함 (Claude 분할 시 사용)
clari task add "제목" --parent 1 --spec "요구사항"
```

### task plan

```bash
# 단일 Task (재귀 분할)
clari task plan [id]

# 전체 todo
clari task plan --all
```

### task run

```bash
# 단일 leaf Task 실행
clari task run [id]

# 전체 planned leaf 실행
clari task run --all
```

---

## DB 스키마

```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER,
    title TEXT NOT NULL,
    spec TEXT DEFAULT '',
    plan TEXT DEFAULT '',
    report TEXT DEFAULT '',
    status TEXT DEFAULT 'todo'
        CHECK(status IN ('todo', 'split', 'planned', 'done', 'failed')),
    error TEXT DEFAULT '',
    is_leaf INTEGER DEFAULT 1,
    depth INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_tasks_leaf ON tasks(is_leaf);
```

---

## 시스템 프롬프트

1회차 순회 시 Claude에게 주입되는 프롬프트:

`bot/internal/prompts/dev.platform/task.md`

---

## 구현 현황

### Phase 1-5: 기존 기능 ✅

### Phase 6: 재귀 분할 시스템 ✅
- [x] Task 구조체에 `IsLeaf`, `Depth` 추가
- [x] DB 스키마에 `is_leaf`, `depth`, `split` 상태 추가
- [x] `task add --spec` 플래그 추가
- [x] 1회차 프롬프트 템플릿 (`task.md`)
- [x] 출력 파서 (`[SPLIT]`, `[PLANNED]`)
- [x] `planRecursive()` 재귀 순회
- [x] 2회차 순회 leaf 조건 추가

---

*Claribot Task System v0.3*
