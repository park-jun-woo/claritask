package test

import (
	"testing"

	"parkjunwoo.com/claritask/internal/service"
)

func TestSetState(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	err := service.SetState(database, "test_key", "test_value")
	if err != nil {
		t.Fatalf("failed to set state: %v", err)
	}

	value, err := service.GetState(database, "test_key")
	if err != nil {
		t.Fatalf("failed to get state: %v", err)
	}

	if value != "test_value" {
		t.Errorf("expected value 'test_value', got '%s'", value)
	}
}

func TestSetStateUpsert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// First set
	service.SetState(database, "key", "value1")

	// Update
	err := service.SetState(database, "key", "value2")
	if err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	value, _ := service.GetState(database, "key")
	if value != "value2" {
		t.Errorf("expected value 'value2', got '%s'", value)
	}
}

func TestGetState(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.SetState(database, "my_key", "my_value")

	value, err := service.GetState(database, "my_key")
	if err != nil {
		t.Fatalf("failed to get state: %v", err)
	}

	if value != "my_value" {
		t.Errorf("expected value 'my_value', got '%s'", value)
	}
}

func TestGetStateNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	value, err := service.GetState(database, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return empty string for non-existent key
	if value != "" {
		t.Errorf("expected empty string, got '%s'", value)
	}
}

func TestGetAllStates(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.SetState(database, "key1", "value1")
	service.SetState(database, "key2", "value2")
	service.SetState(database, "key3", "value3")

	states, err := service.GetAllStates(database)
	if err != nil {
		t.Fatalf("failed to get all states: %v", err)
	}

	if len(states) != 3 {
		t.Errorf("expected 3 states, got %d", len(states))
	}

	if states["key1"] != "value1" {
		t.Errorf("expected key1='value1', got '%s'", states["key1"])
	}
	if states["key2"] != "value2" {
		t.Errorf("expected key2='value2', got '%s'", states["key2"])
	}
	if states["key3"] != "value3" {
		t.Errorf("expected key3='value3', got '%s'", states["key3"])
	}
}

func TestGetAllStatesEmpty(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	states, err := service.GetAllStates(database)
	if err != nil {
		t.Fatalf("failed to get all states: %v", err)
	}

	if len(states) != 0 {
		t.Errorf("expected 0 states, got %d", len(states))
	}
}

func TestDeleteState(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.SetState(database, "to_delete", "value")

	err := service.DeleteState(database, "to_delete")
	if err != nil {
		t.Fatalf("failed to delete state: %v", err)
	}

	// Verify state is deleted
	value, _ := service.GetState(database, "to_delete")
	if value != "" {
		t.Errorf("expected empty string after delete, got '%s'", value)
	}
}

func TestDeleteStateNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	err := service.DeleteState(database, "nonexistent")
	if err == nil {
		t.Error("expected error when deleting non-existent state")
	}
}

func TestStateConstants(t *testing.T) {
	if service.StateCurrentProject != "current_project" {
		t.Errorf("expected StateCurrentProject='current_project', got '%s'", service.StateCurrentProject)
	}
	if service.StateCurrentFeature != "current_feature" {
		t.Errorf("expected StateCurrentFeature='current_feature', got '%s'", service.StateCurrentFeature)
	}
	if service.StateCurrentTask != "current_task" {
		t.Errorf("expected StateCurrentTask='current_task', got '%s'", service.StateCurrentTask)
	}
	if service.StateNextTask != "next_task" {
		t.Errorf("expected StateNextTask='next_task', got '%s'", service.StateNextTask)
	}
}

func TestUpdateCurrentState(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	err := service.UpdateCurrentState(database, "test-project", 1, 10, 11)
	if err != nil {
		t.Fatalf("failed to update current state: %v", err)
	}

	states, _ := service.GetAllStates(database)

	if states["current_project"] != "test-project" {
		t.Errorf("expected current_project='test-project', got '%s'", states["current_project"])
	}
	if states["current_feature"] != "1" {
		t.Errorf("expected current_feature='1', got '%s'", states["current_feature"])
	}
	if states["current_task"] != "10" {
		t.Errorf("expected current_task='10', got '%s'", states["current_task"])
	}
	if states["next_task"] != "11" {
		t.Errorf("expected next_task='11', got '%s'", states["next_task"])
	}
}

func TestUpdateCurrentStateNoNextTask(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	err := service.UpdateCurrentState(database, "test-project", 1, 10, 0)
	if err != nil {
		t.Fatalf("failed to update current state: %v", err)
	}

	value, _ := service.GetState(database, "next_task")
	if value != "" {
		t.Errorf("expected next_task='', got '%s'", value)
	}
}

func TestUpdateCurrentStatePartial(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Only set project
	err := service.UpdateCurrentState(database, "test-project", 0, 0, 0)
	if err != nil {
		t.Fatalf("failed to update current state: %v", err)
	}

	projectValue, _ := service.GetState(database, "current_project")
	if projectValue != "test-project" {
		t.Errorf("expected current_project='test-project', got '%s'", projectValue)
	}

	// Feature should not be set (0 doesn't update)
	featureValue, _ := service.GetState(database, "current_feature")
	if featureValue != "" {
		t.Errorf("expected current_feature='', got '%s'", featureValue)
	}
}

func TestInitializeProjectState(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	err := service.InitializeProjectState(database, "my-project")
	if err != nil {
		t.Fatalf("failed to init state: %v", err)
	}

	states, _ := service.GetAllStates(database)

	if states["current_project"] != "my-project" {
		t.Errorf("expected current_project='my-project', got '%s'", states["current_project"])
	}
	if states["current_phase"] != "" {
		t.Errorf("expected current_phase='', got '%s'", states["current_phase"])
	}
	if states["current_task"] != "" {
		t.Errorf("expected current_task='', got '%s'", states["current_task"])
	}
	if states["next_task"] != "" {
		t.Errorf("expected next_task='', got '%s'", states["next_task"])
	}
}

func TestInitializeProjectStateOverwrite(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Set some existing state
	service.SetState(database, "current_project", "old-project")
	service.SetState(database, "current_task", "5")

	// Init should overwrite
	err := service.InitializeProjectState(database, "new-project")
	if err != nil {
		t.Fatalf("failed to init state: %v", err)
	}

	projectValue, _ := service.GetState(database, "current_project")
	if projectValue != "new-project" {
		t.Errorf("expected current_project='new-project', got '%s'", projectValue)
	}

	taskValue, _ := service.GetState(database, "current_task")
	if taskValue != "" {
		t.Errorf("expected current_task='', got '%s'", taskValue)
	}
}
