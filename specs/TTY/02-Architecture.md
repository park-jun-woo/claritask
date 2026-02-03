# TTY Handover: Architecture

> **버전**: v0.0.1

---

## 전체 구조

```
clari init
  └─▶ Claude Code [Phase 1: 요구사항 수립]
        │
        │  clari feature add '...'
        │  clari feature add '...'
        │
        │  사용자: "개발해"
        │
        └─▶ clari project start
              │
              ├─▶ Claude Code [Task 1] ─▶ 완료
              ├─▶ Claude Code [Task 2] ─▶ 완료
              ├─▶ Claude Code [Task N] ─▶ 완료
              │
              └─▶ 최종 보고
```

---

## TTY Handover 다이어그램

```
┌─────────────────────────────────────────────────────────────────┐
│  Claritask (Orchestrator)                                       │
│                                                                 │
│  TTY Handover ─────────────────────────────────┐                │
│        │                                       │                │
│        ▼                                       ▼                │
│  ┌─────────────────────────────────────────────────────┐        │
│  │  Claude Code (대화형)                                │        │
│  │  - stdin/stdout/stderr 연결                         │        │
│  │  - 사용자 모니터에 표시                              │        │
│  │  - 코딩 → 테스트 → 에러 분석 → 수정 → 반복          │        │
│  │  - 필요시 사용자 키보드 개입 가능                    │        │
│  └─────────────────────────────────────────────────────┘        │
│        │                                                        │
│        ▼ (Claude 종료)                                          │
│  제어권 복귀                                                     │
│        ↓                                                        │
│  다음 단계 진행                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 핵심 설계 원칙

### 1. 순차 실행 (중첩 아님)

```
clari init → Claude Code → (종료)
                              ↓
                        clari project start → Claude Code [Task 1] → (종료)
                                                                       ↓
                                              Claude Code [Task 2] → (종료)
                                                                       ↓
                                                                     ...
```

동시에 쌓이는 게 아니라 **하나 끝나면 다음**. 스택이 깊어지지 않음.

### 2. Stateless 프로세스, State는 DB에

```
┌─────────┐     ┌─────────┐     ┌─────────┐
│  clari  │────▶│   DB    │◀────│  clari  │
└─────────┘     └─────────┘     └─────────┘
     │                               │
     ▼                               ▼
Claude Code                    Claude Code
```

- 각 `clari` 프로세스는 **stateless**
- 모든 상태는 **DB에 저장**
- 어디서든 **재개 가능**

### 3. 실패 복구 용이

```bash
# Phase 1 중간에 크래시
$ clari init  # 다시 실행 → DB 상태 보고 이어서

# Phase 2 중간에 크래시
$ clari project start  # 마지막 완료된 Task 이후부터 재개
```

---

## Phase별 TTY Handover

### Phase 1: 요구사항 수립

```
clari init                     Claude Code
    │                               │
    │  TTY Handover ───────────────▶│
    │                               │ 사용자와 대화
    │                               │ feature add 실행
    │                               │ project start 실행
    │                               │
    │◀─────────────────────────────│ 종료
    │                               │
    └── (종료)
```

### Phase 2: Task 실행

```
clari project start            Claude Code
       │                            │
       │  Task 1 시작                │
       │  TTY Handover ────────────▶│
       │                            │ 코딩
       │                            │ 테스트
       │                            │ 디버깅
       │                            │
       │◀──────────────────────────│ 완료, 종료
       │                            │
       │  Task 2 시작                │
       │  TTY Handover ────────────▶│
       │                            │ ...
```

---

## 프로세스 관계

**`clari → Claude Code → clari → Claude Code` 구조**

이 구조가 가능한 이유:
1. **순차 실행** - 프로세스가 중첩되지 않음
2. **Stateless** - 각 clari는 DB만 참조
3. **독립적** - 각 단계가 독립적으로 재실행 가능

---

*TTY Handover Specification v0.0.1 - 2026-02-03*
