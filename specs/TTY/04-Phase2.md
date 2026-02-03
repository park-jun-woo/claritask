# TTY Handover: Phase 2 - 자동 실행

> **버전**: v0.0.1

---

## 목적

확정된 Features를 코드로 구현합니다.

---

## 실행 방법

```bash
$ clari project start
# → Task 목록 생성 → Edge 연결 → 순차 실행
```

---

## 단계

```
1. Plan
   └─ Features → FDL 상세화 → 스켈레톤 생성

2. Task List Up
   └─ 스켈레톤 기반 Task 목록 생성

3. Task Edge Link
   └─ Task 간 의존성 그래프 구성

4. Task 실행 (반복)
   ├─ TTY Handover → Claude Code에게 완전 위임
   ├─ 코딩 + 테스트 + 디버깅
   ├─ Task 완료 → Claude Code 종료 → Claritask로 복귀
   └─ 다음 Task로 이동

5. 최종 보고
   └─ 전체 실행 결과 요약
```

---

## 플로우

```
┌─────────────────────────────────────────────────────────────┐
│  Phase 2: 자동 실행                                          │
│                                                             │
│  $ clari project start  (새 프로세스)                        │
│       ↓                                                     │
│  Plan → Task List → Edge Link                               │
│       ↓                                                     │
│  ┌─────────────────────────────────────────┐                │
│  │  Task N 실행                             │                │
│  │  ┌─────────────────────────────────────┐│                │
│  │  │  TTY Handover → Claude Code         ││  ← 완전 위임   │
│  │  │  (코딩 + 테스트 + 디버깅)            ││                │
│  │  │  완료 → Claude Code 종료            ││                │
│  │  └─────────────────────────────────────┘│                │
│  └─────────────────────────────────────────┘                │
│       ↓ (반복)                                               │
│  최종 보고                                                   │
└─────────────────────────────────────────────────────────────┘
```

---

## TTY Handover (Task 단위)

```
clari project start                Claude Code
       │                                │
       │  Task 1 시작                    │
       │  TTY Handover ────────────────▶│
       │                                │ 코딩
       │                                │ 테스트
       │                                │ 디버깅
       │                                │
       │◀──────────────────────────────│ 완료, 종료
       │                                │
       │  Task 2 시작                    │
       │  TTY Handover ────────────────▶│
       │                                │ ...
       │                                │
       │◀──────────────────────────────│ 완료, 종료
       │                                │
       │  ...                           │
       │                                │
       └── 최종 보고
```

---

## Claude Code에게 전달되는 정보

각 Task 실행 시 Claude Code에게 전달:

```
[CLARITASK TASK SESSION]

Task ID: 42
Target File: services/comment_service.py
Target Function: createComment
Test Command: pytest test_comment.py::test_create

=== FDL Specification ===
service:
  - name: createComment
    input: { userId, postId, content }
    steps: [validate, db insert, return]

=== Current Skeleton Code ===
async def createComment(...):
    # TODO: implement
    raise NotImplementedError

=== Dependency Results ===
Task 41 (Comment model): 완료 - models/comment.py 생성됨

---
Start by running the test command.
```

---

## 실패 처리

### 코드 레벨 실패 (테스트 실패, 컴파일 에러)

1. Claude Code가 자체 디버깅 시도
2. 해결되면 계속
3. 안 되면 Task failed → 다음 Task로 진행 또는 중단

### 요구사항 레벨 실패 (스펙 모호, 결정 필요)

1. Task failed 처리
2. `clari project start` 중단
3. 사용자가 스펙 수정 후 재실행

### 재개

```bash
# 실패 후 재개
$ clari project start
# → 마지막 성공한 Task 이후부터 재개
```

---

## 사후 검증

각 Task 완료 후 Claritask가 검증:

```go
func VerifyAfterTask(task Task) (bool, error) {
    // 테스트 재실행
    cmd := exec.Command("sh", "-c", task.TestCmd)
    output, err := cmd.CombinedOutput()

    if err == nil {
        return true, nil  // 성공
    }
    return false, err     // 실패
}
```

---

## 최종 보고

모든 Task 완료 후:

```
[Claritask] Project Execution Complete

=== Summary ===
Total Tasks: 15
Completed: 14
Failed: 1

=== Failed Tasks ===
- Task 12: payment_api (reason: external API timeout)

=== Generated Files ===
- models/user.py
- models/product.py
- services/auth_service.py
- services/product_service.py
- ...

Total execution time: 45m 32s
```

---

*TTY Handover Specification v0.0.1 - 2026-02-03*
