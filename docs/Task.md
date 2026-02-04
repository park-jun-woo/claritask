# Task 시스템 설계

> Claribot의 Task 기반 분할 정복 시스템

---

## 개요

대규모 작업을 여러 Task로 분할하고, 연관 Task의 컨텍스트를 주입하여 일관성 있는 실행을 보장한다.

**핵심 아이디어**: Edge 방향을 무시하고, 순회 단계로 데이터 흐름을 제어한다.

---

## Task 구조

```go
Task {
    ID          int
    ParentID    *int      // 상위 Task (연결의 일종)
    Title       string    // 제목
    Spec        string    // 요구사항 명세서 (사용자 입력, 불변)
    Plan        string    // 계획서 (1회차 순회에서 생성)
    Report      string    // 보고서 (2회차 순회 후 생성)
    Status      string    // spec_ready → plan_ready → done
    CreatedAt   string
    UpdatedAt   string
}
```

### 필드 설명

| 필드 | 설명 | 생성 시점 |
|------|------|----------|
| Title | 작업 제목 | 등록 시 |
| Spec | 요구사항 명세서 (불변) | 등록 시 |
| Plan | 실행 계획서 | 1회차 순회 |
| Report | 완료 보고서 | 2회차 순회 |

### 상태 전이

```
spec_ready → plan_ready → done
    ↑            ↑          ↑
  등록 시    1회차 완료  2회차 완료
```

---

## 연결 모델

### 단순화된 연결

- **Edge**: Task 간 연결 (방향 무시)
- **Parent-Child**: 상위-하위 관계 (연결의 일종)

```
연결 있음 = Edge 존재 OR ParentID 참조

방향성 → 무시
관계 타입 → 구분 안 함
```

### 연관 Task 조회

```sql
-- Edge로 연결된 Task
SELECT * FROM tasks WHERE id IN (
    SELECT to_task_id FROM task_edges WHERE from_task_id = ?
    UNION
    SELECT from_task_id FROM task_edges WHERE to_task_id = ?
)

-- Parent/Child 관계
SELECT * FROM tasks WHERE id = ? OR parent_id = ?
```

---

## 실행 흐름

### 전체 흐름

```
[등록] → [1회차 순회] → [2회차 순회]
           (plan)        (execute)
```

### 상세 흐름

```
[등록]
    └─ title + spec 입력
    └─ status = 'spec_ready'

[1회차 순회 - Plan 생성]
    └─ 대상: status = 'spec_ready'
    └─ 입력: 본인 spec + 연관 tasks의 spec
    └─ 출력: plan
    └─ 완료: status = 'plan_ready'

[2회차 순회 - 실행]
    └─ 대상: status = 'plan_ready'
    └─ 입력: 본인 plan + 연관 tasks의 plan
    └─ 출력: 코드 실행 + report
    └─ 완료: status = 'done'
```

### 데이터 흐름 다이어그램

```
┌─────────────────────────────────────────────────────────┐
│                      [등록 단계]                         │
│  Task A: spec_a    Task B: spec_b    Task C: spec_c    │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                    [1회차 순회]                          │
│                                                         │
│  Task A: spec_a + 연관 specs → plan_a                   │
│  Task B: spec_b + 연관 specs → plan_b                   │
│  Task C: spec_c + 연관 specs → plan_c                   │
│                                                         │
│  (모든 Task의 plan 생성 완료 후 다음 단계)                │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                    [2회차 순회]                          │
│                                                         │
│  Task A: plan_a + 연관 plans → 실행 → report_a          │
│  Task B: plan_b + 연관 plans → 실행 → report_b          │
│  Task C: plan_c + 연관 plans → 실행 → report_c          │
└─────────────────────────────────────────────────────────┘
```

---

## 프롬프트 구조

### 1회차 (Plan 생성)

```markdown
# Task: {title}

## 요구사항
{spec}

## 연관 자료

### Task #{id}: {title}
**명세서**: {spec}

### Task #{id}: {title}
**명세서**: {spec}

---

위 요구사항과 연관 자료를 참고하여 실행 계획서를 작성하세요.

계획서에는 다음 내용을 포함하세요:
- 구현 방향
- 주요 변경 파일
- 의존성 또는 주의사항
```

### 2회차 (실행)

```markdown
# Task: {title}

## 계획서
{plan}

## 연관 자료

### Task #{id}: {title}
**계획서**: {plan}

### Task #{id}: {title}
**계획서**: {plan}

---

위 계획서와 연관 자료를 참고하여 작업을 수행하세요.

완료 후 보고서를 작성하세요:
- 수행한 작업 요약
- 변경된 파일 목록
- 특이사항
```

---

## 핵심 설계 원칙

1. **Edge 방향 무시**: 연결만 있으면 연관 자료에 포함
2. **순회 단계로 흐름 제어**: 1회차 전체 완료 → 2회차 시작
3. **Spec 불변**: 원래 요구사항은 수정하지 않음
4. **단계별 데이터 격리**:
   - 1회차: spec만 참조
   - 2회차: plan만 참조

---

## DB 스키마 변경

### 기존

```sql
tasks (
    id, parent_id, source, title, content, status, result, error,
    created_at, started_at, completed_at
)
```

### 변경

```sql
tasks (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER,
    title TEXT,
    spec TEXT,           -- 요구사항 명세서
    plan TEXT,           -- 계획서
    report TEXT,         -- 완료 보고서
    status TEXT,         -- 'spec_ready', 'plan_ready', 'done', 'failed'
    error TEXT,
    created_at TEXT,
    updated_at TEXT
)
```

### 삭제된 필드

- `source`: 불필요 (Task는 항상 내부 생성)
- `content`: `spec`으로 대체
- `result`: `report`로 대체
- `started_at`, `completed_at`: `updated_at`으로 통합

---

## 구현 TODO

### Phase 1: 스키마 변경
- [ ] tasks 테이블 마이그레이션
- [ ] Task 구조체 수정
- [ ] CRUD 함수 수정

### Phase 2: 1회차 순회 (Plan 생성)
- [ ] 연관 Task 조회 함수
- [ ] Plan 생성 프롬프트
- [ ] 순회 실행 함수

### Phase 3: 2회차 순회 (실행)
- [ ] 실행 프롬프트
- [ ] Claude Code 연동
- [ ] Report 저장

### Phase 4: CLI/Telegram 연동
- [ ] `task plan` 명령어 (1회차 수동 실행)
- [ ] `task run` 명령어 (2회차 수동 실행)
- [ ] `task cycle` 명령어 (1회차 + 2회차 자동)

---

*Claribot Task System v0.1*
