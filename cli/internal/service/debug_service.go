package service

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// RunInteractiveDebugging starts an interactive debugging session
func RunInteractiveDebugging(database *db.DB, task *model.Task) error {
	fmt.Println("[Claritask] Entering Interactive Debugging Mode...")
	fmt.Printf("   Task: %s\n", task.ID)
	fmt.Printf("   Title: %s\n", task.Title)
	if task.TargetFile != "" {
		fmt.Printf("   Target: %s\n", task.TargetFile)
	}
	fmt.Println("   Claude Code will take over. You can intervene if needed.")
	fmt.Println()

	// Build system prompt
	systemPrompt := buildDebugSystemPrompt()

	// Build initial prompt
	initialPrompt := buildDebugInitialPrompt(database, task)

	// Run Claude in interactive mode
	cmd := exec.Command("claude",
		"--system-prompt", systemPrompt,
		"--permission-mode", "acceptEdits",
		initialPrompt,
	)

	// TTY Handover
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute (blocking)
	err := cmd.Run()

	fmt.Println()
	fmt.Println("[Claritask] Debugging Session Ended.")

	return err
}

// buildDebugSystemPrompt creates the system prompt for debugging
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

// buildDebugInitialPrompt creates the initial prompt for debugging
func buildDebugInitialPrompt(database *db.DB, task *model.Task) string {
	var fdlSpec string
	var skeletonCode string

	// Get FDL if feature is associated
	if task.FeatureID > 0 {
		feature, err := GetFeature(database, task.FeatureID)
		if err == nil && feature != nil {
			fdlSpec = feature.FDL
		}
	}

	// Get skeleton content if available
	if task.TargetFile != "" {
		content, err := ReadSkeletonContent(task.TargetFile)
		if err == nil {
			skeletonCode = content
		}
	}

	// Infer test command
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

// inferTestCommand infers the appropriate test command for a task
func inferTestCommand(task *model.Task) string {
	if task.TargetFile == "" {
		return "# Run appropriate test command"
	}

	switch {
	case strings.HasSuffix(task.TargetFile, ".py"):
		testFile := strings.Replace(task.TargetFile, ".py", "_test.py", 1)
		testFile = strings.Replace(testFile, "/", "/test_", 1)
		return fmt.Sprintf("pytest %s -v", testFile)
	case strings.HasSuffix(task.TargetFile, ".go"):
		dir := strings.TrimSuffix(task.TargetFile, "/"+task.TargetFile[strings.LastIndex(task.TargetFile, "/")+1:])
		return fmt.Sprintf("go test %s -v", dir)
	case strings.HasSuffix(task.TargetFile, ".ts") || strings.HasSuffix(task.TargetFile, ".tsx"):
		return "npm test"
	case strings.HasSuffix(task.TargetFile, ".js") || strings.HasSuffix(task.TargetFile, ".jsx"):
		return "npm test"
	default:
		return "# Run appropriate test command"
	}
}

// VerifyAfterDebugging verifies the fix after debugging
func VerifyAfterDebugging(task *model.Task) (bool, error) {
	fmt.Println("[Claritask] Verifying fix...")

	testCmd := inferTestCommand(task)
	if testCmd == "# Run appropriate test command" {
		// Cannot auto-verify
		return true, nil
	}

	cmd := exec.Command("sh", "-c", testCmd)
	output, err := cmd.CombinedOutput()

	if err == nil {
		fmt.Println("Verification Passed!")
		return true, nil
	}

	fmt.Println("Verification Failed.")
	fmt.Printf("Output:\n%s\n", string(output))
	return false, fmt.Errorf("verification failed: %s", string(output))
}

// ExecuteWithFallback executes a task with fallback to interactive mode
func ExecuteWithFallback(database *db.DB, task *model.Task, manifest *model.Manifest) error {
	// Try headless execution first
	result, err := ExecuteTaskWithClaude(task, manifest)
	if err == nil && result.Success {
		return nil
	}

	fmt.Println("Headless execution failed. Switching to interactive mode...")

	// Fallback to interactive mode
	if err := RunInteractiveDebugging(database, task); err != nil {
		return fmt.Errorf("interactive debugging failed: %w", err)
	}

	// Verify after debugging
	passed, err := VerifyAfterDebugging(task)
	if !passed {
		return fmt.Errorf("verification failed after debugging: %w", err)
	}

	return nil
}

// MaxDebugAttempts is the maximum number of debugging attempts
const MaxDebugAttempts = 3
