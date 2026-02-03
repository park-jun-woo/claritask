package test

import (
	"os"
	"path/filepath"
	"testing"

	"parkjunwoo.com/claritask/internal/service"
)

func TestAddExpert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create experts directory
	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	// Test adding expert
	expert, err := service.AddExpert(database, "test-expert")
	if err != nil {
		t.Fatalf("failed to add expert: %v", err)
	}

	if expert.ID != "test-expert" {
		t.Errorf("expected ID 'test-expert', got '%s'", expert.ID)
	}

	// Verify file was created
	expertPath := filepath.Join(".claritask/experts", "test-expert", "EXPERT.md")
	if _, err := os.Stat(expertPath); os.IsNotExist(err) {
		t.Errorf("EXPERT.md file was not created")
	}
}

func TestAddExpertDuplicate(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	// Add first expert
	_, err := service.AddExpert(database, "dup-expert")
	if err != nil {
		t.Fatalf("failed to add first expert: %v", err)
	}

	// Try to add duplicate
	_, err = service.AddExpert(database, "dup-expert")
	if err == nil {
		t.Error("expected error for duplicate expert, got nil")
	}
}

func TestAddExpertInvalidID(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	// Test invalid IDs
	invalidIDs := []string{
		"Test-Expert",  // uppercase
		"test expert",  // space
		"test_expert",  // underscore
		"-test",        // starts with hyphen
		"test-",        // ends with hyphen
	}

	for _, id := range invalidIDs {
		_, err := service.AddExpert(database, id)
		if err == nil {
			t.Errorf("expected error for invalid ID '%s', got nil", id)
		}
	}
}

func TestListExperts(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	// Add experts
	service.AddExpert(database, "expert-a")
	service.AddExpert(database, "expert-b")

	// List all
	experts, err := service.ListExperts(database, "all")
	if err != nil {
		t.Fatalf("failed to list experts: %v", err)
	}

	if len(experts) != 2 {
		t.Errorf("expected 2 experts, got %d", len(experts))
	}
}

func TestGetExpert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	// Add expert
	service.AddExpert(database, "get-test")

	// Get expert
	expert, err := service.GetExpert(database, "get-test")
	if err != nil {
		t.Fatalf("failed to get expert: %v", err)
	}

	if expert.ID != "get-test" {
		t.Errorf("expected ID 'get-test', got '%s'", expert.ID)
	}
}

func TestGetExpertNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	_, err := service.GetExpert(database, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent expert, got nil")
	}
}

func TestRemoveExpert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	// Add expert
	service.AddExpert(database, "remove-test")

	// Remove expert
	err := service.RemoveExpert(database, "remove-test", false)
	if err != nil {
		t.Fatalf("failed to remove expert: %v", err)
	}

	// Verify removal
	_, err = service.GetExpert(database, "remove-test")
	if err == nil {
		t.Error("expected error after removal, got nil")
	}
}

func TestAssignExpert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	// Create project
	service.CreateProject(database, "test-proj", "Test Project", "")

	// Add expert
	service.AddExpert(database, "assign-test")

	// Assign expert
	err := service.AssignExpert(database, "test-proj", "assign-test")
	if err != nil {
		t.Fatalf("failed to assign expert: %v", err)
	}

	// Verify assignment
	expert, _ := service.GetExpert(database, "assign-test")
	if !expert.Assigned {
		t.Error("expected expert to be assigned")
	}
}

func TestAssignExpertDuplicate(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	service.CreateProject(database, "test-proj", "Test Project", "")
	service.AddExpert(database, "assign-dup")
	service.AssignExpert(database, "test-proj", "assign-dup")

	// Try duplicate assignment
	err := service.AssignExpert(database, "test-proj", "assign-dup")
	if err == nil {
		t.Error("expected error for duplicate assignment, got nil")
	}
}

func TestUnassignExpert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	service.CreateProject(database, "test-proj", "Test Project", "")
	service.AddExpert(database, "unassign-test")
	service.AssignExpert(database, "test-proj", "unassign-test")

	// Unassign
	err := service.UnassignExpert(database, "test-proj", "unassign-test")
	if err != nil {
		t.Fatalf("failed to unassign expert: %v", err)
	}

	// Verify
	expert, _ := service.GetExpert(database, "unassign-test")
	if expert.Assigned {
		t.Error("expected expert to be unassigned")
	}
}

func TestGetAssignedExperts(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	service.CreateProject(database, "test-proj", "Test Project", "")
	service.AddExpert(database, "assigned-a")
	service.AddExpert(database, "assigned-b")
	service.AssignExpert(database, "test-proj", "assigned-a")
	service.AssignExpert(database, "test-proj", "assigned-b")

	experts, err := service.GetAssignedExperts(database, "test-proj")
	if err != nil {
		t.Fatalf("failed to get assigned experts: %v", err)
	}

	if len(experts) != 2 {
		t.Errorf("expected 2 assigned experts, got %d", len(experts))
	}

	// Check content is included
	for _, e := range experts {
		if e.Content == "" {
			t.Errorf("expected content for expert %s, got empty", e.ID)
		}
	}
}

func TestRemoveAssignedExpert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	service.CreateProject(database, "test-proj", "Test Project", "")
	service.AddExpert(database, "remove-assigned")
	service.AssignExpert(database, "test-proj", "remove-assigned")

	// Try to remove without force
	err := service.RemoveExpert(database, "remove-assigned", false)
	if err == nil {
		t.Error("expected error when removing assigned expert without force")
	}

	// Remove with force
	err = service.RemoveExpert(database, "remove-assigned", true)
	if err != nil {
		t.Fatalf("failed to force remove expert: %v", err)
	}
}

func TestListExpertsFilter(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.MkdirAll(".claritask/experts", 0755)
	defer os.RemoveAll(".claritask")

	service.CreateProject(database, "test-proj", "Test Project", "")
	service.AddExpert(database, "filter-assigned")
	service.AddExpert(database, "filter-available")
	service.AssignExpert(database, "test-proj", "filter-assigned")

	// Test assigned filter
	assigned, _ := service.ListExperts(database, "assigned")
	if len(assigned) != 1 || assigned[0].ID != "filter-assigned" {
		t.Error("assigned filter failed")
	}

	// Test available filter
	available, _ := service.ListExperts(database, "available")
	if len(available) != 1 || available[0].ID != "filter-available" {
		t.Error("available filter failed")
	}
}
