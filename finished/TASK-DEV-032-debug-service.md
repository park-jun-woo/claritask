# TASK-DEV-032: Debug 서비스

## 개요
- **파일**: `internal/service/debug_service.go`
- **유형**: 신규
- **우선순위**: Medium
- **Phase**: 4 (TTY Handover)
- **예상 LOC**: ~250

## 목적
TTY Handover 기반 대화형 디버깅 모드 구현

## 작업 내용

### 1. 대화형 디버깅 실행

```go
// RunInteractiveDebugging - 대화형 디버깅 모드 실행
func RunInteractiveDebugging(db *db.DB, task *model.Task) error {
    fmt.Println("🚧 [Claritask] Entering Interactive Debugging Mode...")
    fmt.Printf("   Task: %s\n", task.ID)
    fmt.Printf("   Target: %s\n", task.TargetFile)
    fmt.Println("   Claude Code will take over. You can intervene if needed.")
    fmt.Println()

    // 1. 시스템 프롬프트 구성
    systemPrompt := buildDebugSystemPrompt()

    // 2. 초기 프롬프트 구성
    initialPrompt := buildDebugInitialPrompt(db, task)

    // 3. Claude 실행 (대화형 모드)
    cmd := exec.Command("claude",
        "--system-prompt", systemPrompt,
        "--permission-mode", "acceptEdits",
        initialPrompt,
    )

    // 4. TTY 핸드오버
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 5. 실행 (Blocking)
    err := cmd.Run()

    fmt.Println()
    fmt.Println("✅ [Claritask] Debugging Session Ended.")

    return err
}
```

### 2. 디버깅 시스템 프롬프트

```go
func buildDebugSystemPrompt() string {
    return `You are in Claritask Interactive Debugging Mode.

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
If you cannot fix it after 3 attempts, explain the blocker and exit.

IMPORTANT: Start working immediately without waiting for user input.`
}
```

### 3. 디버깅 초기 프롬프트

```go
func buildDebugInitialPrompt(db *db.DB, task *model.Task) string {
    // FDL 조회
    var fdlSpec string
    if task.FeatureID != nil {
        feature, _ := GetFeature(db, *task.FeatureID)
        if feature != nil {
            fdlSpec = feature.FDL
        }
    }

    // Skeleton 조회
    var skeletonCode string
    if task.TargetFile != "" {
        content, _ := ReadFile(task.TargetFile)
        skeletonCode = content
    }

    // 테스트 명령어 추론
    testCmd := inferTestCommand(task)

    return fmt.Sprintf(`[CLARITASK DEBUGGING SESSION]

Task ID: %s
Target File: %s
Target Function: %s
Test Command: %s

=== FDL Specification ===
%s

=== Current Code ===
%s

---
Start by running the test command: %s
`,
        task.ID,
        task.TargetFile,
        task.TargetFunction,
        testCmd,
        fdlSpec,
        skeletonCode,
        testCmd,
    )
}
```

### 4. 테스트 명령어 추론

```go
// inferTestCommand - Task에 맞는 테스트 명령어 추론
func inferTestCommand(task *model.Task) string {
    // 파일 확장자 기반 추론
    switch {
    case strings.HasSuffix(task.TargetFile, ".py"):
        return fmt.Sprintf("pytest %s", getTestFile(task.TargetFile))
    case strings.HasSuffix(task.TargetFile, ".go"):
        return fmt.Sprintf("go test %s", getTestFile(task.TargetFile))
    case strings.HasSuffix(task.TargetFile, ".ts"):
        return "npm test"
    default:
        return "# Run appropriate test command"
    }
}
```

### 5. 사후 검증

```go
// VerifyAfterDebugging - 디버깅 후 테스트 검증
func VerifyAfterDebugging(task *model.Task) (bool, error) {
    fmt.Println("🔍 [Claritask] Verifying fix...")

    testCmd := inferTestCommand(task)
    cmd := exec.Command("sh", "-c", testCmd)
    output, err := cmd.CombinedOutput()

    if err == nil {
        fmt.Println("🎉 Verification Passed!")
        return true, nil
    }

    fmt.Println("⚠️ Verification Failed.")
    fmt.Printf("Output:\n%s\n", string(output))
    return false, fmt.Errorf("verification failed: %s", output)
}
```

### 6. Task 실패 시 Fallback

```go
// ExecuteWithFallback - 비대화형 실행 실패 시 대화형으로 전환
func ExecuteWithFallback(db *db.DB, task *model.Task, manifest *model.Manifest) error {
    // 1. 비대화형 실행 시도
    result, err := ExecuteTaskWithClaude(task, manifest)
    if err == nil && result.Success {
        return nil
    }

    fmt.Println("⚠️ Headless execution failed. Switching to interactive mode...")

    // 2. 대화형 모드로 전환
    if err := RunInteractiveDebugging(db, task); err != nil {
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

### 7. 타임아웃 및 세션 관리

```go
// RunInteractiveDebuggingWithTimeout - 타임아웃 포함 디버깅
func RunInteractiveDebuggingWithTimeout(db *db.DB, task *model.Task, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    // ... 디버깅 로직 (context 사용)
}

// MaxDebugAttempts - 최대 디버깅 시도 횟수
const MaxDebugAttempts = 3
```

## 의존성
- TASK-DEV-029 (Task 서비스 확장)
- TASK-DEV-030 (Orchestrator 서비스)

## 완료 기준
- [ ] TTY Handover 구현됨
- [ ] 시스템 프롬프트 구성됨
- [ ] 초기 프롬프트 구성됨
- [ ] 사후 검증 구현됨
- [ ] Fallback 로직 구현됨
- [ ] go build 성공
- [ ] 수동 테스트 통과
