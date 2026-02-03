# TTY Handover: Implementation

> **버전**: v0.0.1

---

## Go 구현

### 핵심 함수: RunWithTTYHandover

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

func RunWithTTYHandover(task Task, contextPacket string) error {
    fmt.Println("🚀 [Claritask] Starting Task:", task.ID)
    fmt.Printf("   Target: %s\n", task.TargetFile)
    fmt.Println()

    // 1. 시스템 프롬프트 구성
    systemPrompt := buildSystemPrompt(task)

    // 2. 초기 프롬프트 구성
    initialPrompt := buildInitialPrompt(task, contextPacket)

    // 3. Claude 실행 (대화형 모드)
    cmd := exec.Command("claude",
        "--system-prompt", systemPrompt,
        "--permission-mode", "acceptEdits",
        initialPrompt,
    )

    // 4. TTY 핸드오버: 터미널 입출력 연결
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 5. 실행 및 대기 (Blocking)
    err := cmd.Run()

    fmt.Println()
    fmt.Println("✅ [Claritask] Task Session Ended.")

    return err
}
```

### 시스템 프롬프트 빌더

```go
func buildSystemPrompt(task Task) string {
    return `You are in Claritask Task Execution Mode.

ROLE: Implement the TODO section in the target file.

WORKFLOW:
1. Read the target file
2. Implement the TODO section following the FDL specification
3. Run the test command
4. If test fails, analyze and fix
5. Repeat until test passes

CONSTRAINTS:
- Do NOT modify function signatures (generated from FDL)
- Only implement the TODO sections
- Follow the FDL specification exactly

COMPLETION:
When the test passes, summarize what you implemented and exit with /exit.
If you cannot complete after 3 attempts, explain the blocker and exit.

IMPORTANT: Start working immediately without waiting for user input.`
}
```

### 초기 프롬프트 빌더

```go
func buildInitialPrompt(task Task, contextPacket string) string {
    return fmt.Sprintf(`[CLARITASK TASK SESSION]

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
Start by reading the target file and implementing the TODO section.
Then run: %s
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
}
```

---

## 사후 검증

```go
func VerifyAfterTask(task Task) (bool, error) {
    fmt.Println("🔍 [Claritask] Verifying...")

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

---

## 전체 Task 실행 플로우

```go
func ExecuteTask(task Task, context string) error {
    // 1. TTY Handover로 Claude Code 실행
    if err := RunWithTTYHandover(task, context); err != nil {
        return fmt.Errorf("task execution failed: %w", err)
    }

    // 2. 사후 검증
    passed, err := VerifyAfterTask(task)
    if !passed {
        return fmt.Errorf("verification failed: %w", err)
    }

    // 3. Task 완료 처리
    if err := MarkTaskComplete(task.ID); err != nil {
        return fmt.Errorf("failed to mark task complete: %w", err)
    }

    return nil
}
```

---

## CLI 명령어

### 자동 실행

```bash
# 전체 프로젝트 실행
clari project start

# 특정 Feature만 실행
clari project start --feature 2

# Dry-run (실행 없이 Task 목록만)
clari project start --dry-run
```

### 수동 Task 실행

```bash
# 특정 Task 실행
clari task run <task_id>

# 실패한 Task 재시도
clari task retry <task_id>
```

### 실행 중단/재개

```bash
# 실행 중단
clari project stop

# 상태 확인
clari project status

# 재개 (마지막 성공 Task 이후부터)
clari project start
```

---

## 고려사항

### 1. 타임아웃

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

cmd := exec.CommandContext(ctx, "claude", ...)
```

### 2. 로깅

```go
// Task 세션 로그 저장
logFile, _ := os.Create(fmt.Sprintf(".claritask/logs/task-%s.log", task.ID))
cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
```

### 3. 최대 시도 횟수

```go
const MaxAttempts = 3

systemPrompt += fmt.Sprintf(`
MAX ATTEMPTS: %d
If you cannot complete after %d attempts, exit and report the blocker.
`, MaxAttempts, MaxAttempts)
```

---

*TTY Handover Specification v0.0.1 - 2026-02-03*
