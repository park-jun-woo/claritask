# TTY Handover: 대화형 디버깅 모드

## 개요

Claritask가 `claude --print`로 비대화형 실행 중 디버깅이 필요한 상황에서, **터미널 제어권을 Claude Code에게 잠시 넘겨주고(Foreground), Claude가 종료되면 다시 제어권을 가져오는 방식**.

---

## 아키텍처

```
┌─────────────────────────────────────────────────────────────────┐
│  Claritask (Orchestrator)                                       │
│                                                                 │
│  평소: claude --print (Headless, 비대화형)                       │
│        ↓                                                        │
│  테스트 실패 감지                                                │
│        ↓                                                        │
│  TTY Handover ─────────────────────────────────┐                │
│        │                                       │                │
│        ▼                                       ▼                │
│  ┌─────────────────────────────────────────────────────┐        │
│  │  Claude Code (대화형)                                │        │
│  │  - stdin/stdout/stderr 연결                         │        │
│  │  - 사용자 모니터에 표시                              │        │
│  │  - 테스트 → 에러 분석 → 코드 수정 → 반복            │        │
│  │  - 필요시 사용자 키보드 개입 가능                    │        │
│  └─────────────────────────────────────────────────────┘        │
│        │                                                        │
│        ▼ (Claude 종료)                                          │
│  제어권 복귀 + 사후 검증                                         │
│        ↓                                                        │
│  다음 Task 진행                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Claude CLI 옵션

### 프롬프트 전달

```bash
# 대화형 모드 + 첫 프롬프트 (positional argument)
claude "테스트 실행하고 버그 고쳐줘"

# 시스템 프롬프트 + 첫 프롬프트
claude --system-prompt "너는 디버깅 전문가야" "pytest 실행해"
```

### 권한 모드

```bash
--permission-mode <mode>

# 옵션:
# - default: 기본 (도구 실행 전 확인)
# - acceptEdits: 편집 자동 승인
# - bypassPermissions: 모든 권한 확인 건너뛰기
# - dontAsk: 묻지 않음
# - plan: 계획 모드
```

### 세션 관리

```bash
--continue              # 가장 최근 대화 이어가기
--resume <session_id>   # 특정 세션 복원
--session-id <uuid>     # 특정 세션 ID 사용
```

---

## Go 구현

### 핵심 함수: RunInteractiveDebugging

```go
package orchestrator

import (
    "fmt"
    "os"
    "os/exec"
)

type Task struct {
    ID           string
    TargetFile   string
    TargetFunc   string
    TestCmd      string
    FDL          string
    SkeletonCode string
}

func RunInteractiveDebugging(task Task, contextPacket string) error {
    fmt.Println("🚧 [Claritask] Entering Interactive Debugging Mode...")
    fmt.Printf("   Task: %s\n", task.ID)
    fmt.Printf("   Target: %s\n", task.TargetFile)
    fmt.Println("   Claude Code will take over. You can intervene if needed.")
    fmt.Println()

    // 1. 시스템 프롬프트 구성
    systemPrompt := `You are in Claritask Interactive Debugging Mode.

ROLE: Debug and fix failing tests autonomously.

WORKFLOW:
1. Run the test command
2. Analyze the error output
3. Read the relevant code
4. Edit the code to fix the issue
5. Run the test again
6. Repeat until the test passes

CONSTRAINTS:
- Do NOT modify function signatures (they are generated from FDL)
- Only implement the TODO sections
- Follow the FDL specification exactly

COMPLETION:
When the test passes, summarize what you fixed and exit with /exit.
If you cannot fix it after 3 attempts, explain the blocker and exit.`

    // 2. 초기 프롬프트 구성
    initialPrompt := fmt.Sprintf(`[CLARITASK DEBUGGING SESSION]

Task ID: %s
Target File: %s
Target Function: %s
Test Command: %s

=== FDL Specification ===
%s

=== Current Skeleton Code ===
%s

=== Additional Context ===
%s

---
Start by running the test command: %s
`,
        task.ID,
        task.TargetFile,
        task.TargetFunc,
        task.TestCmd,
        task.FDL,
        task.SkeletonCode,
        contextPacket,
        task.TestCmd,
    )

    // 3. Claude 실행 (대화형 모드)
    cmd := exec.Command("claude",
        "--system-prompt", systemPrompt,
        "--permission-mode", "acceptEdits",  // 편집 자동 승인
        initialPrompt,
    )

    // 4. TTY 핸드오버: 터미널 입출력 연결
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 5. 실행 및 대기 (Blocking)
    // Claude가 종료될 때까지 Claritask는 여기서 대기
    err := cmd.Run()

    fmt.Println()
    fmt.Println("✅ [Claritask] Debugging Session Ended.")

    return err
}
```

### 사후 검증

```go
func VerifyAfterDebugging(task Task) (bool, error) {
    fmt.Println("🔍 [Claritask] Verifying fix...")

    // 테스트 재실행
    cmd := exec.Command("sh", "-c", task.TestCmd)
    output, err := cmd.CombinedOutput()

    if err == nil {
        fmt.Println("🎉 Verification Passed!")
        return true, nil
    }

    fmt.Println("⚠️ Verification Failed.")
    fmt.Printf("Output:\n%s\n", string(output))
    return false, err
}
```

### 전체 플로우

```go
func ExecuteTaskWithFallback(task Task, context string) error {
    // 1. 먼저 비대화형으로 시도
    result, err := RunHeadless(task, context)
    if err == nil {
        return nil
    }

    // 2. 실패 시 대화형 디버깅으로 전환
    fmt.Println("⚠️ Headless execution failed. Switching to interactive mode...")

    if err := RunInteractiveDebugging(task, context); err != nil {
        return fmt.Errorf("interactive debugging failed: %w", err)
    }

    // 3. 사후 검증
    passed, err := VerifyAfterDebugging(task)
    if !passed {
        return fmt.Errorf("verification failed after debugging: %w", err)
    }

    return nil
}
```

---

## CLI 명령어

### 수동 트리거

```bash
# 특정 Task를 대화형으로 실행
clari task debug <task_id>

# 실패한 Task를 대화형으로 재시도
clari task retry <task_id> --interactive
```

### 자동 트리거 설정

```bash
# project start 시 실패하면 자동으로 대화형 전환
clari project start --fallback-interactive

# 또는 설정으로
clari config set debug.auto_interactive true
```

---

## 프롬프트 전략

### Auto-Pilot Trigger

Claude가 대화형 모드에서 "무엇을 도와드릴까요?" 하고 대기하지 않고, **즉시 작업을 시작**하게 하려면:

```text
[시스템 프롬프트 끝에]

IMPORTANT: Start working immediately without waiting for user input.
Your first action should be running the test command.
```

### 컨텍스트 압축

대화형 모드에서도 컨텍스트가 너무 크면 문제. 핵심만 전달:

```text
=== FDL (핵심만) ===
service:
  - name: createComment
    input: { userId, postId, content }
    steps: [validate, db insert, return]

=== 에러 로그 (최근 50줄) ===
...

=== 관련 코드 (TODO 부분만) ===
async def createComment(...):
    # TODO: implement
    raise NotImplementedError
```

---

## 고려사항

### 1. 무한 루프 방지

```go
const MaxDebugAttempts = 3

systemPrompt += fmt.Sprintf(`
MAX ATTEMPTS: %d
If you cannot fix after %d attempts, exit and report the blocker.
`, MaxDebugAttempts, MaxDebugAttempts)
```

### 2. 타임아웃

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

cmd := exec.CommandContext(ctx, "claude", ...)
```

### 3. 세션 저장

```bash
# 디버깅 세션을 나중에 복원할 수 있도록
claude --session-id <uuid> ...

# 나중에 이어서
claude --resume <uuid>
```

### 4. 로깅

```go
// 디버깅 세션 로그 저장
logFile, _ := os.Create(fmt.Sprintf(".claritask/debug-logs/%s.log", task.ID))
cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
```

---

## 사용 시나리오

### 시나리오 1: 자동 실행 중 실패

```
[Claritask] Executing Task 42: createComment...
[Claritask] Running: pytest test_comment.py::test_create
[Claritask] ❌ Test failed. Switching to interactive mode...

🚧 [Claritask] Entering Interactive Debugging Mode...
   Task: 42
   Target: services/comment_service.py
   Claude Code will take over.

╭─────────────────────────────────────────────╮
│  Claude Code                                │
│                                             │
│  > Running pytest test_comment.py::test_... │
│  > Error: ValidationError at line 23        │
│  > Reading services/comment_service.py...   │
│  > Editing line 23-25...                    │
│  > Running pytest again...                  │
│  > ✓ Test passed!                           │
│                                             │
│  Fixed: Added content length validation.    │
│  /exit                                      │
╰─────────────────────────────────────────────╯

✅ [Claritask] Debugging Session Ended.
🔍 [Claritask] Verifying fix...
🎉 Verification Passed!
[Claritask] Task 42 completed. Moving to Task 43...
```

### 시나리오 2: 사용자 개입 필요

```
╭─────────────────────────────────────────────╮
│  Claude Code                                │
│                                             │
│  > Error: Missing environment variable      │
│  > DB_CONNECTION_STRING not set             │
│                                             │
│  I need the database connection string.     │
│  Please provide it or set the env variable. │
│                                             │
│  User: export DB_CONNECTION_STRING=...      │ ← 사용자 개입
│                                             │
│  > Retrying...                              │
│  > ✓ Test passed!                           │
╰─────────────────────────────────────────────╯
```

---

## 요약

| 항목 | 설명 |
|------|------|
| **목적** | 비대화형 실행 실패 시 대화형으로 전환하여 디버깅 |
| **방식** | TTY 핸드오버 (stdin/stdout/stderr 연결) |
| **트리거** | 테스트 실패, 또는 수동 `clari task debug` |
| **권한** | `--permission-mode acceptEdits` |
| **복귀** | Claude 종료 시 자동 제어권 복귀 |
| **검증** | 핸드오버 종료 후 Claritask가 테스트 재실행 |

**"평소에는 자동, 막히면 수동"** - 두 모드의 장점을 모두 활용.
