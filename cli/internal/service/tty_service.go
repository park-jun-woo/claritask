package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	return `You are in Claritask FDL Generation Mode.

ROLE: Generate a Feature Definition Language (FDL) YAML file based on the feature description.

FDL STRUCTURE:
- feature: Feature name (snake_case)
- description: Feature description
- models: Data layer (database tables/models)
- service: Logic layer (business logic functions)
- api: Interface layer (API endpoints)
- ui: Presentation layer (UI components)

IMPORTANT:
1. Generate a complete FDL YAML file
2. Save the FDL to: features/<feature-name>.fdl.yaml
3. After saving, update the DB: clari fdl register features/<feature-name>.fdl.yaml
4. Exit with /exit when done

CONSTRAINTS:
- Follow the 4-layer structure (models, service, api, ui)
- Use snake_case for feature names and function names
- Be specific in field definitions and API contracts`
}

// RunFDLGenerationWithTTY runs FDL generation with TTY handover
func RunFDLGenerationWithTTY(database *db.DB, featureID int64, featureName, description string) error {
	fmt.Println("[Claritask] Starting FDL Generation")
	fmt.Printf("   Feature: %s (ID: %d)\n", featureName, featureID)
	fmt.Println("   Claude Code will generate FDL specification.")
	fmt.Println()

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

	err := RunWithTTYHandover(systemPrompt, initialPrompt, "acceptEdits")

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

Please generate a complete FDL YAML file for this feature.

Output file: features/%s.fdl.yaml

After generating, save the file and run:
clari fdl register features/%s.fdl.yaml

Then exit with /exit.
`, featureID, name, description, projectContext, techContext, designContext, name, name)
}
