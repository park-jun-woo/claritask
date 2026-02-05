package task

import (
	"os"
	"path/filepath"
	"testing"

	"parkjunwoo.com/claribot/internal/db"
)

func setupTestDB(t *testing.T) (string, func()) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "claribot-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create .claribot directory
	claribotDir := filepath.Join(tmpDir, ".claribot")
	if err := os.MkdirAll(claribotDir, 0755); err != nil {
		t.Fatalf("Failed to create .claribot dir: %v", err)
	}

	// Open and migrate DB
	localDB, err := db.OpenLocal(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	if err := localDB.MigrateLocal(); err != nil {
		t.Fatalf("Failed to migrate DB: %v", err)
	}
	localDB.Close()

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestTaskAdd(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	result := Add(projectPath, "Test Task", nil, "Test spec")
	if !result.Success {
		t.Errorf("Add failed: %s", result.Message)
	}

	task, ok := result.Data.(*Task)
	if !ok {
		t.Fatal("Expected Task data")
	}

	if task.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", task.Title)
	}

	if task.Status != "spec_ready" {
		t.Errorf("Expected status 'spec_ready', got '%s'", task.Status)
	}

	if task.Spec != "Test spec" {
		t.Errorf("Expected spec 'Test spec', got '%s'", task.Spec)
	}
}

func TestTaskAddWithParent(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	// Add parent task
	parentResult := Add(projectPath, "Parent Task", nil, "")
	if !parentResult.Success {
		t.Fatalf("Add parent failed: %s", parentResult.Message)
	}

	parent := parentResult.Data.(*Task)

	// Add child task
	childResult := Add(projectPath, "Child Task", &parent.ID, "")
	if !childResult.Success {
		t.Errorf("Add child failed: %s", childResult.Message)
	}

	child := childResult.Data.(*Task)
	if child.ParentID == nil || *child.ParentID != parent.ID {
		t.Error("Child task should have parent ID")
	}

	// Check depth
	if child.Depth != 1 {
		t.Errorf("Expected child depth 1, got %d", child.Depth)
	}
}

func TestTaskGet(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	// Add task
	addResult := Add(projectPath, "Test Task", nil, "")
	if !addResult.Success {
		t.Fatalf("Add failed: %s", addResult.Message)
	}

	task := addResult.Data.(*Task)

	// Get task
	getResult := Get(projectPath, "1")
	if !getResult.Success {
		t.Errorf("Get failed: %s", getResult.Message)
	}

	gotTask := getResult.Data.(*Task)
	if gotTask.ID != task.ID {
		t.Errorf("Expected ID %d, got %d", task.ID, gotTask.ID)
	}
}

func TestTaskSet(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	// Add task
	Add(projectPath, "Test Task", nil, "")

	// Set spec
	result := Set(projectPath, "1", "spec", "Test specification")
	if !result.Success {
		t.Errorf("Set spec failed: %s", result.Message)
	}

	// Verify
	getResult := Get(projectPath, "1")
	task := getResult.Data.(*Task)
	if task.Spec != "Test specification" {
		t.Errorf("Expected spec 'Test specification', got '%s'", task.Spec)
	}
}

func TestTaskSetStatus(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	Add(projectPath, "Test Task", nil, "")

	// Set invalid status
	result := Set(projectPath, "1", "status", "invalid")
	if result.Success {
		t.Error("Expected failure for invalid status")
	}

	// Set valid status
	result = Set(projectPath, "1", "status", "plan_ready")
	if !result.Success {
		t.Errorf("Set status failed: %s", result.Message)
	}
}

func TestBuildPlanPrompt(t *testing.T) {
	task := &Task{
		ID:    1,
		Title: "Test Task",
		Spec:  "Test specification",
	}

	related := []Task{
		{ID: 2, Title: "Related Task", Spec: "Related spec"},
	}

	prompt := BuildPlanPrompt(task, related, "/tmp/test-report.md")

	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	// Check contains task info
	if !contains(prompt, "Test Task") {
		t.Error("Prompt should contain task title")
	}

	if !contains(prompt, "Test specification") {
		t.Error("Prompt should contain task spec")
	}

	// Check contains related info
	if !contains(prompt, "Related Task") {
		t.Error("Prompt should contain related task title")
	}
}

func TestBuildExecutePrompt(t *testing.T) {
	task := &Task{
		ID:    1,
		Title: "Test Task",
		Plan:  "Test plan",
	}

	related := []Task{
		{ID: 2, Title: "Related Task", Plan: "Related plan"},
	}

	prompt := BuildExecutePrompt(task, related, "/tmp/test-report.md")

	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	if !contains(prompt, "Test plan") {
		t.Error("Prompt should contain task plan")
	}

	if !contains(prompt, "Related plan") {
		t.Error("Prompt should contain related task plan")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
