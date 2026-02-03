# TASK-TTY-001: TTY Service 통합

## 개요
TTY Handover 관련 함수를 tty_service.go로 통합

## 배경
- **스펙**: specs/TTY/05-Implementation.md
- **현재 상태**: debug_service.go에 부분 구현

## 작업 내용

### 1. tty_service.go 생성
**파일**: `cli/internal/service/tty_service.go`

```go
package service

import (
    "fmt"
    "os"
    "os/exec"
    "parkjunwoo.com/claritask/internal/db"
    "parkjunwoo.com/claritask/internal/model"
)

// RunWithTTYHandover executes Claude with TTY handover
func RunWithTTYHandover(systemPrompt, initialPrompt string, permissionMode string) error {
    args := []string{}

    if systemPrompt != "" {
        args = append(args, "--system-prompt", systemPrompt)
    }
    if permissionMode != "" {
        args = append(args, "--permission-mode", permissionMode)
    }
    args = append(args, initialPrompt)

    cmd := exec.Command("claude", args...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}

// Phase1SystemPrompt returns system prompt for requirements gathering
func Phase1SystemPrompt(projectID, projectName string) string {
    return fmt.Sprintf(`You are in Claritask Phase 1: Requirements Gathering Mode.

PROJECT: %s (%s)

ROLE: Help the user define project features through conversation.

WORKFLOW:
1. Understand the user's project idea
2. Propose initial features
3. Refine based on user feedback
4. Save features using: clari feature add '{"name":"...", "description":"..."}'

COMPLETION:
When the user says "개발해", "시작해", or "만들어줘":
1. Confirm all features are saved
2. Run: clari project start
3. Exit with /exit

IMPORTANT: Start by asking about the project requirements.`, projectName, projectID)
}

// Phase2SystemPrompt returns system prompt for task execution
func Phase2SystemPrompt() string {
    return `You are in Claritask Phase 2: Task Execution Mode.

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

// BuildTaskPromptForTTY builds initial prompt for TTY task execution
func BuildTaskPromptForTTY(database *db.DB, task *model.Task) string {
    var fdlSpec, skeletonCode, testCmd string

    if task.FeatureID > 0 {
        if feature, err := GetFeature(database, task.FeatureID); err == nil {
            fdlSpec = feature.FDL
        }
    }

    if task.TargetFile != "" {
        if content, err := ReadSkeletonContent(task.TargetFile); err == nil {
            skeletonCode = content
        }
    }

    testCmd = InferTestCommand(task.TargetFile)

    return fmt.Sprintf(`[CLARITASK TASK SESSION]

Task ID: %d
Title: %s
Target File: %s
Target Function: %s
Test Command: %s

=== Task Content ===
%s

=== FDL Specification ===
%s

=== Current Code ===
%s

---
Start by running: %s
`, task.ID, task.Title, task.TargetFile, task.TargetFunction, testCmd,
   task.Content, fdlSpec, skeletonCode, testCmd)
}
```

### 2. InferTestCommand 함수 export
debug_service.go의 inferTestCommand를 InferTestCommand로 export

## 완료 기준
- [ ] tty_service.go 생성
- [ ] Phase1SystemPrompt 함수
- [ ] Phase2SystemPrompt 함수
- [ ] RunWithTTYHandover 함수
- [ ] BuildTaskPromptForTTY 함수
