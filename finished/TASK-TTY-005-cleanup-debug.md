# TASK-TTY-005: debug_service.go 정리

## 개요
debug_service.go의 중복 코드를 tty_service.go로 이전하고 정리

## 배경
- debug_service.go에 TTY 관련 코드가 있음
- tty_service.go로 통합 후 정리 필요

## 작업 내용

### 1. debug_service.go에서 제거할 함수
- `buildDebugSystemPrompt()` → Phase2SystemPrompt() 사용
- `buildDebugInitialPrompt()` → BuildTaskPromptForTTY() 사용
- `inferTestCommand()` → InferTestCommand()로 export

### 2. debug_service.go 수정
**파일**: `cli/internal/service/debug_service.go`

```go
package service

import (
    "parkjunwoo.com/claritask/internal/db"
    "parkjunwoo.com/claritask/internal/model"
)

// RunInteractiveDebugging starts an interactive debugging session
// Deprecated: Use RunTaskWithTTY instead
func RunInteractiveDebugging(database *db.DB, task *model.Task) error {
    return RunTaskWithTTY(database, task)
}

// VerifyAfterDebugging verifies the fix after debugging
// Deprecated: Use VerifyTask instead
func VerifyAfterDebugging(task *model.Task) (bool, error) {
    return VerifyTask(task)
}

// ExecuteWithFallback executes a task with fallback to interactive mode
func ExecuteWithFallback(database *db.DB, task *model.Task, manifest *model.Manifest) error {
    // Try headless execution first
    result, err := ExecuteTaskWithClaude(task, manifest)
    if err == nil && result.Success {
        return nil
    }

    // Fallback to interactive mode
    if err := RunTaskWithTTY(database, task); err != nil {
        return err
    }

    // Verify after debugging
    passed, err := VerifyTask(task)
    if !passed {
        return err
    }

    return nil
}

// MaxDebugAttempts is the maximum number of debugging attempts
const MaxDebugAttempts = 3
```

### 3. InferTestCommand export
**파일**: `cli/internal/service/tty_service.go`

```go
// InferTestCommand infers the appropriate test command for a file
func InferTestCommand(targetFile string) string {
    if targetFile == "" {
        return ""
    }

    switch {
    case strings.HasSuffix(targetFile, ".py"):
        testFile := strings.Replace(targetFile, ".py", "_test.py", 1)
        return fmt.Sprintf("pytest %s -v", testFile)
    case strings.HasSuffix(targetFile, ".go"):
        dir := filepath.Dir(targetFile)
        return fmt.Sprintf("go test %s -v", dir)
    case strings.HasSuffix(targetFile, ".ts"), strings.HasSuffix(targetFile, ".tsx"):
        return "npm test"
    case strings.HasSuffix(targetFile, ".js"), strings.HasSuffix(targetFile, ".jsx"):
        return "npm test"
    default:
        return ""
    }
}
```

## 완료 기준
- [ ] debug_service.go 코드 정리
- [ ] InferTestCommand export
- [ ] 하위 호환성 유지 (deprecated 함수 래핑)
