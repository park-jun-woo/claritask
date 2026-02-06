package task

import (
	"testing"

	"parkjunwoo.com/claribot/internal/db"
)

func TestGetRelatedParentChild(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	// Add parent and child
	parentResult := Add(projectPath, "Parent", nil, "")
	parent := parentResult.Data.(*Task)

	Add(projectPath, "Child", &parent.ID, "")

	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer localDB.Close()

	// Get related from parent's perspective (should include child)
	relatedFromParent, err := GetRelated(localDB, 1)
	if err != nil {
		t.Fatalf("GetRelated failed: %v", err)
	}

	if len(relatedFromParent) != 1 {
		t.Errorf("Expected 1 related task (child), got %d", len(relatedFromParent))
	}

	// Get related from child's perspective (should include parent)
	relatedFromChild, err := GetRelated(localDB, 2)
	if err != nil {
		t.Fatalf("GetRelated failed: %v", err)
	}

	if len(relatedFromChild) != 1 {
		t.Errorf("Expected 1 related task (parent), got %d", len(relatedFromChild))
	}
}

func TestBuildContextMap(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	Add(projectPath, "Task 1", nil, "")
	Add(projectPath, "Task 2", nil, "")

	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer localDB.Close()

	contextMap, err := BuildContextMap(localDB)
	if err != nil {
		t.Fatalf("BuildContextMap failed: %v", err)
	}

	if contextMap == "" {
		t.Error("Expected non-empty context map")
	}

	// Should contain task IDs
	if !contains(contextMap, "#1") {
		t.Error("Context map should contain #1")
	}
	if !contains(contextMap, "#2") {
		t.Error("Context map should contain #2")
	}
}
