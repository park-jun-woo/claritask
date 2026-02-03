package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/docs"
	"parkjunwoo.com/claritask/internal/model"
)

// Session management
var (
	sessionMutex   sync.Mutex
	activeSessions int
	sessionCond    *sync.Cond
)

func init() {
	sessionCond = sync.NewCond(&sessionMutex)
}

// acquireSession waits until a session slot is available
func acquireSession(maxSessions int) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	for activeSessions >= maxSessions {
		fmt.Printf("[Claritask] Waiting for session slot... (%d/%d active)\n", activeSessions, maxSessions)
		sessionCond.Wait()
	}

	activeSessions++
	fmt.Printf("[Claritask] Session acquired (%d/%d active)\n", activeSessions, maxSessions)
}

// releaseSession releases a session slot
func releaseSession() {
	sessionMutex.Lock()
	activeSessions--
	fmt.Printf("[Claritask] Session released (%d active)\n", activeSessions)
	sessionMutex.Unlock()

	sessionCond.Signal()
}

// GetSessionStatus returns current session status
func GetSessionStatus() (active, max int) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	config, _ := LoadConfig()
	return activeSessions, config.TTY.MaxParallelSessions
}

// RunWithTTYHandover executes Claude with TTY handover
func RunWithTTYHandover(systemPrompt, initialPrompt string, permissionMode string) error {
	return RunWithTTYHandoverEx(systemPrompt, initialPrompt, permissionMode, "")
}

// RunWithTTYHandoverEx executes Claude with TTY handover and optional completion file watching
func RunWithTTYHandoverEx(systemPrompt, initialPrompt string, permissionMode string, completeFile string) error {
	// Load config and acquire session slot
	config, _ := LoadConfig()
	acquireSession(config.TTY.MaxParallelSessions)
	defer releaseSession()

	args := []string{"--dangerously-skip-permissions"}

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

	if err := cmd.Start(); err != nil {
		return err
	}

	// If completion file is specified, watch for it
	if completeFile != "" {
		go watchCompleteFile(completeFile, cmd)
	}

	return cmd.Wait()
}

// watchCompleteFile watches for completion file and terminates the process
func watchCompleteFile(completeFile string, cmd *exec.Cmd) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if cmd.ProcessState != nil {
			// Process already finished
			return
		}

		if _, err := os.Stat(completeFile); err == nil {
			// Complete file exists, terminate Claude
			fmt.Println("\n[Claritask] Completion detected. Closing session...")

			if cmd.Process != nil {
				cmd.Process.Kill()
			}

			// Delete the complete file
			os.Remove(completeFile)
			return
		}
	}
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

AVAILABLE COMMANDS:
- clari feature add '{"name":"feature-name", "description":"Feature description"}' - Add a feature
- clari feature list - List all features
- clari project start - Start execution (Phase 2)

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

// RunTaskWithTTY runs a single task with TTY handover
func RunTaskWithTTY(database *db.DB, task *model.Task) error {
	fmt.Printf("[Claritask] Running Task %d: %s\n", task.ID, task.Title)
	if task.TargetFile != "" {
		fmt.Printf("   Target: %s\n", task.TargetFile)
	}
	fmt.Println("   Claude Code will take over. You can intervene if needed.")
	fmt.Println()

	systemPrompt := Phase2SystemPrompt()
	initialPrompt := BuildTaskPromptForTTY(database, task)

	err := RunWithTTYHandover(systemPrompt, initialPrompt, "acceptEdits")

	fmt.Println()
	fmt.Println("[Claritask] Task Session Ended.")

	return err
}

// BuildTaskPromptForTTY builds initial prompt for TTY task execution
func BuildTaskPromptForTTY(database *db.DB, task *model.Task) string {
	var fdlSpec, skeletonCode string

	if task.FeatureID > 0 {
		if feature, err := GetFeature(database, task.FeatureID); err == nil && feature != nil {
			fdlSpec = feature.FDL
		}
	}

	if task.TargetFile != "" {
		if content, err := ReadSkeletonContent(task.TargetFile); err == nil {
			skeletonCode = content
		}
	}

	testCmd := InferTestCommand(task.TargetFile)

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

// InferTestCommand infers the appropriate test command for a file
func InferTestCommand(targetFile string) string {
	if targetFile == "" {
		return "# Run appropriate test command"
	}

	switch {
	case strings.HasSuffix(targetFile, ".py"):
		testFile := strings.Replace(targetFile, ".py", "_test.py", 1)
		return fmt.Sprintf("pytest %s -v", testFile)
	case strings.HasSuffix(targetFile, ".go"):
		dir := filepath.Dir(targetFile)
		if dir == "." || dir == "" {
			dir = "./..."
		}
		return fmt.Sprintf("go test %s -v", dir)
	case strings.HasSuffix(targetFile, ".ts"), strings.HasSuffix(targetFile, ".tsx"):
		return "npm test"
	case strings.HasSuffix(targetFile, ".js"), strings.HasSuffix(targetFile, ".jsx"):
		return "npm test"
	default:
		return "# Run appropriate test command"
	}
}

// VerifyTask verifies task completion by running test command
func VerifyTask(task *model.Task) (bool, error) {
	testCmd := InferTestCommand(task.TargetFile)
	if testCmd == "" || testCmd == "# Run appropriate test command" {
		// Cannot auto-verify
		return true, nil
	}

	fmt.Printf("[Claritask] Verifying: %s\n", testCmd)

	cmd := exec.Command("sh", "-c", testCmd)
	output, err := cmd.CombinedOutput()

	if err == nil {
		fmt.Println("[Claritask] Verification passed!")
		return true, nil
	}

	return false, fmt.Errorf("verification failed: %s", string(output))
}

// RunInteractiveInit starts interactive requirements gathering
func RunInteractiveInit(database *db.DB, projectID, projectName, description string) error {
	fmt.Println("[Claritask] Starting Phase 1: Requirements Gathering")
	fmt.Printf("   Project: %s (%s)\n", projectName, projectID)
	fmt.Println("   Claude Code will help you define features.")
	fmt.Println()

	systemPrompt := Phase1SystemPrompt(projectID, projectName)
	initialPrompt := fmt.Sprintf(`프로젝트: %s
설명: %s

위 프로젝트에 필요한 기능들을 함께 정의해봅시다. 어떤 기능이 필요한지 알려주세요.`, projectName, description)

	err := RunWithTTYHandover(systemPrompt, initialPrompt, "")

	fmt.Println()
	fmt.Println("[Claritask] Phase 1 Session Ended.")

	return err
}

// FDLGenerationSystemPrompt returns system prompt for FDL generation
func FDLGenerationSystemPrompt() string {
	return fmt.Sprintf(`You are in Claritask FDL Generation Mode.

ROLE: Generate a Feature Definition Language (FDL) YAML file based on the feature description.

=== FDL SPECIFICATION ===
%s
=== END FDL SPECIFICATION ===

WORKFLOW:
1. Analyze the feature description
2. Design the 4-layer structure (models, service, api, ui)
3. Generate a complete FDL YAML file
4. Save to: features/<feature-name>.fdl.yaml

COMPLETION:
When FDL file is saved, create an empty file: .claritask/complete
This signals that your work is done.

CONSTRAINTS:
- Follow the FDL specification exactly
- Use snake_case for feature names
- Use camelCase for service function names
- Use PascalCase for model and component names
- All api must have 'use:' field (wiring to service)
- All ui actions must have API path`, docs.FDLSpec)
}

// RunFDLGenerationWithTTY runs FDL generation with TTY handover
func RunFDLGenerationWithTTY(database *db.DB, featureID int64, featureName, description string) error {
	fmt.Println("[Claritask] Starting FDL Generation")
	fmt.Printf("   Feature: %s (ID: %d)\n", featureName, featureID)
	fmt.Println("   Claude Code will generate FDL specification.")
	fmt.Println()

	// Ensure .claritask directory exists
	claritaskDir := ".claritask"
	if err := os.MkdirAll(claritaskDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claritask directory: %w", err)
	}

	// Remove any existing complete file
	completeFile := filepath.Join(claritaskDir, "complete")
	os.Remove(completeFile)

	// Get project context
	var projectContext, techContext, designContext string
	if ctx, err := GetContext(database); err == nil && ctx != nil {
		projectContext = fmt.Sprintf("%v", ctx)
	}
	if tech, err := GetTech(database); err == nil && tech != nil {
		techContext = fmt.Sprintf("%v", tech)
	}
	if design, err := GetDesign(database); err == nil && design != nil {
		designContext = fmt.Sprintf("%v", design)
	}

	systemPrompt := FDLGenerationSystemPrompt()
	initialPrompt := BuildFDLPrompt(featureID, featureName, description, projectContext, techContext, designContext)

	err := RunWithTTYHandoverEx(systemPrompt, initialPrompt, "acceptEdits", completeFile)

	// Cleanup complete file if it exists
	os.Remove(completeFile)

	fmt.Println()
	fmt.Println("[Claritask] FDL Generation Session Ended.")

	return err
}

// BuildFDLPrompt builds the initial prompt for FDL generation
func BuildFDLPrompt(featureID int64, name, description, projectContext, techContext, designContext string) string {
	return fmt.Sprintf(`[CLARITASK FDL GENERATION]

Feature ID: %d
Feature Name: %s
Description: %s

=== Project Context ===
%s

=== Tech Stack ===
%s

=== Design Decisions ===
%s

---

Generate a complete FDL YAML file for this feature.

Output file: features/%s.fdl.yaml

IMPORTANT: After saving the FDL file, create an empty file '.claritask/complete' to signal completion.
Example: touch .claritask/complete
`, featureID, name, description, projectContext, techContext, designContext, name)
}
