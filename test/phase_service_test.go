package test

import (
	"testing"

	"parkjunwoo.com/talos/internal/service"
)

func TestCreatePhase(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create project first
	err := service.CreateProject(database, "test-project", "Test Project", "Description")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	input := service.PhaseCreateInput{
		ProjectID:   "test-project",
		Name:        "Phase 1",
		Description: "First phase",
		OrderNum:    1,
	}

	id, err := service.CreatePhase(database, input)
	if err != nil {
		t.Fatalf("failed to create phase: %v", err)
	}

	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCreatePhaseInvalidProject(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	input := service.PhaseCreateInput{
		ProjectID:   "non-existent",
		Name:        "Phase 1",
		Description: "First phase",
		OrderNum:    1,
	}

	// This should fail due to foreign key constraint
	_, err := service.CreatePhase(database, input)
	if err == nil {
		t.Error("expected error when creating phase with invalid project ID")
	}
}

func TestGetPhase(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	input := service.PhaseCreateInput{
		ProjectID:   "test-project",
		Name:        "Phase 1",
		Description: "First phase",
		OrderNum:    1,
	}
	id, _ := service.CreatePhase(database, input)

	phase, err := service.GetPhase(database, id)
	if err != nil {
		t.Fatalf("failed to get phase: %v", err)
	}

	if phase.Name != "Phase 1" {
		t.Errorf("expected Name 'Phase 1', got '%s'", phase.Name)
	}
	if phase.Description != "First phase" {
		t.Errorf("expected Description 'First phase', got '%s'", phase.Description)
	}
	if phase.OrderNum != 1 {
		t.Errorf("expected OrderNum 1, got %d", phase.OrderNum)
	}
	if phase.Status != "pending" {
		t.Errorf("expected Status 'pending', got '%s'", phase.Status)
	}
	if phase.ProjectID != "test-project" {
		t.Errorf("expected ProjectID 'test-project', got '%s'", phase.ProjectID)
	}
}

func TestGetPhaseNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := service.GetPhase(database, 999)
	if err == nil {
		t.Error("expected error when getting non-existent phase")
	}
}

func TestListPhases(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	// Create phases out of order
	service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 3",
		OrderNum:  3,
	})
	service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})
	service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 2",
		OrderNum:  2,
	})

	phases, err := service.ListPhases(database, "test-project")
	if err != nil {
		t.Fatalf("failed to list phases: %v", err)
	}

	if len(phases) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(phases))
	}

	// Should be ordered by order_num
	if phases[0].Name != "Phase 1" {
		t.Errorf("expected first phase 'Phase 1', got '%s'", phases[0].Name)
	}
	if phases[1].Name != "Phase 2" {
		t.Errorf("expected second phase 'Phase 2', got '%s'", phases[1].Name)
	}
	if phases[2].Name != "Phase 3" {
		t.Errorf("expected third phase 'Phase 3', got '%s'", phases[2].Name)
	}
}

func TestListPhasesEmpty(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")

	phases, err := service.ListPhases(database, "test-project")
	if err != nil {
		t.Fatalf("failed to list phases: %v", err)
	}

	if len(phases) != 0 {
		t.Errorf("expected 0 phases, got %d", len(phases))
	}
}

func TestUpdatePhaseStatus(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	id, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})

	err := service.UpdatePhaseStatus(database, id, "active")
	if err != nil {
		t.Fatalf("failed to update phase status: %v", err)
	}

	phase, _ := service.GetPhase(database, id)
	if phase.Status != "active" {
		t.Errorf("expected Status 'active', got '%s'", phase.Status)
	}
}

func TestStartPhase(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	id, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})

	err := service.StartPhase(database, id)
	if err != nil {
		t.Fatalf("failed to start phase: %v", err)
	}

	phase, _ := service.GetPhase(database, id)
	if phase.Status != "active" {
		t.Errorf("expected Status 'active', got '%s'", phase.Status)
	}
}

func TestStartPhaseAlreadyActive(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	id, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})

	// Start once
	service.StartPhase(database, id)

	// Try to start again
	err := service.StartPhase(database, id)
	if err == nil {
		t.Error("expected error when starting already active phase")
	}
}

func TestCompletePhase(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	id, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})

	// Must start first
	service.StartPhase(database, id)

	err := service.CompletePhase(database, id)
	if err != nil {
		t.Fatalf("failed to complete phase: %v", err)
	}

	phase, _ := service.GetPhase(database, id)
	if phase.Status != "done" {
		t.Errorf("expected Status 'done', got '%s'", phase.Status)
	}
}

func TestCompletePhaseNotActive(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	id, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})

	// Try to complete without starting
	err := service.CompletePhase(database, id)
	if err == nil {
		t.Error("expected error when completing non-active phase")
	}
}

func TestListPhasesWithTaskCount(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	phaseID, _ := service.CreatePhase(database, service.PhaseCreateInput{
		ProjectID: "test-project",
		Name:      "Phase 1",
		OrderNum:  1,
	})

	// Create tasks
	taskInput1 := service.TaskCreateInput{
		PhaseID: phaseID,
		Title:   "Task 1",
	}
	taskInput2 := service.TaskCreateInput{
		PhaseID: phaseID,
		Title:   "Task 2",
	}
	taskInput3 := service.TaskCreateInput{
		PhaseID: phaseID,
		Title:   "Task 3",
	}

	service.CreateTask(database, taskInput1)
	taskID2, _ := service.CreateTask(database, taskInput2)
	service.CreateTask(database, taskInput3)

	// Complete one task
	service.StartTask(database, taskID2)
	service.CompleteTask(database, taskID2, "done")

	phases, err := service.ListPhases(database, "test-project")
	if err != nil {
		t.Fatalf("failed to list phases: %v", err)
	}

	if len(phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(phases))
	}

	if phases[0].TasksTotal != 3 {
		t.Errorf("expected TasksTotal 3, got %d", phases[0].TasksTotal)
	}
	if phases[0].TasksDone != 1 {
		t.Errorf("expected TasksDone 1, got %d", phases[0].TasksDone)
	}
}
