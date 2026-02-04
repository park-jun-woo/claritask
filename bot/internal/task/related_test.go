package task

import (
	"testing"

	"parkjunwoo.com/claribot/internal/db"
)

func TestGetRelated(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	// Add tasks
	Add(projectPath, "Task 1", nil)
	Add(projectPath, "Task 2", nil)
	Add(projectPath, "Task 3", nil)

	// Set specs
	Set(projectPath, "1", "spec", "Spec 1")
	Set(projectPath, "2", "spec", "Spec 2")

	// Open DB for edge creation
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer localDB.Close()

	// Add edge between task 1 and task 2
	now := db.TimeNow()
	_, err = localDB.Exec(`INSERT INTO task_edges (from_task_id, to_task_id, created_at) VALUES (?, ?, ?)`, 1, 2, now)
	if err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	// Get related tasks for task 1
	related, err := GetRelated(localDB, 1)
	if err != nil {
		t.Fatalf("GetRelated failed: %v", err)
	}

	if len(related) != 1 {
		t.Errorf("Expected 1 related task, got %d", len(related))
	}

	if related[0].ID != 2 {
		t.Errorf("Expected related task ID 2, got %d", related[0].ID)
	}
}

func TestGetRelatedBidirectional(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	Add(projectPath, "Task 1", nil)
	Add(projectPath, "Task 2", nil)

	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer localDB.Close()

	// Add edge from 1 to 2
	now := db.TimeNow()
	localDB.Exec(`INSERT INTO task_edges (from_task_id, to_task_id, created_at) VALUES (?, ?, ?)`, 1, 2, now)

	// Get related from task 2's perspective (reverse direction)
	related, err := GetRelated(localDB, 2)
	if err != nil {
		t.Fatalf("GetRelated failed: %v", err)
	}

	if len(related) != 1 {
		t.Errorf("Expected 1 related task (bidirectional), got %d", len(related))
	}

	if related[0].ID != 1 {
		t.Errorf("Expected related task ID 1, got %d", related[0].ID)
	}
}

func TestGetRelatedParentChild(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	// Add parent and child
	parentResult := Add(projectPath, "Parent", nil)
	parent := parentResult.Data.(*Task)

	Add(projectPath, "Child", &parent.ID)

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

func TestGetRelatedSpecs(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	Add(projectPath, "Task 1", nil)
	Add(projectPath, "Task 2", nil)
	Add(projectPath, "Task 3", nil)

	// Only set spec for task 2
	Set(projectPath, "2", "spec", "Spec for task 2")

	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer localDB.Close()

	// Add edges
	now := db.TimeNow()
	localDB.Exec(`INSERT INTO task_edges (from_task_id, to_task_id, created_at) VALUES (?, ?, ?)`, 1, 2, now)
	localDB.Exec(`INSERT INTO task_edges (from_task_id, to_task_id, created_at) VALUES (?, ?, ?)`, 1, 3, now)

	// Get related specs (should only return task 2 which has spec)
	relatedSpecs, err := GetRelatedSpecs(localDB, 1)
	if err != nil {
		t.Fatalf("GetRelatedSpecs failed: %v", err)
	}

	if len(relatedSpecs) != 1 {
		t.Errorf("Expected 1 task with spec, got %d", len(relatedSpecs))
	}

	if relatedSpecs[0].ID != 2 {
		t.Errorf("Expected task ID 2, got %d", relatedSpecs[0].ID)
	}
}

func TestGetRelatedPlans(t *testing.T) {
	projectPath, cleanup := setupTestDB(t)
	defer cleanup()

	Add(projectPath, "Task 1", nil)
	Add(projectPath, "Task 2", nil)

	// Set plan for task 2
	Set(projectPath, "2", "plan", "Plan for task 2")

	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer localDB.Close()

	// Add edge
	now := db.TimeNow()
	localDB.Exec(`INSERT INTO task_edges (from_task_id, to_task_id, created_at) VALUES (?, ?, ?)`, 1, 2, now)

	// Get related plans
	relatedPlans, err := GetRelatedPlans(localDB, 1)
	if err != nil {
		t.Fatalf("GetRelatedPlans failed: %v", err)
	}

	if len(relatedPlans) != 1 {
		t.Errorf("Expected 1 task with plan, got %d", len(relatedPlans))
	}
}
