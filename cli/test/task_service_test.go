package test

import (
	"testing"

	"parkjunwoo.com/claritask/internal/service"
)

func TestCreateTask(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	input := service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Test Task",
		Content:   "Task content",
	}

	taskID, err := service.CreateTask(database, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if taskID <= 0 {
		t.Errorf("expected positive task ID, got %d", taskID)
	}
}

func TestGetTask(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	input := service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Test Task",
		Content:   "Task content",
	}
	taskID, _ := service.CreateTask(database, input)

	task, err := service.GetTask(database, taskID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if task.Title != "Test Task" {
		t.Errorf("expected Title 'Test Task', got '%s'", task.Title)
	}
	if task.Content != "Task content" {
		t.Errorf("expected Content 'Task content', got '%s'", task.Content)
	}
	if task.Status != "pending" {
		t.Errorf("expected Status 'pending', got '%s'", task.Status)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := service.GetTask(database, 999)
	if err == nil {
		t.Error("expected error when getting non-existent task")
	}
}

func TestListTasksByFeature(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	// Create multiple tasks
	service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 1"})
	service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 2"})
	service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 3"})

	tasks, err := service.ListTasksByFeature(database, featureID)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestStartTask(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	taskID, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Task 1",
	})

	err := service.StartTask(database, taskID)
	if err != nil {
		t.Fatalf("failed to start task: %v", err)
	}

	task, _ := service.GetTask(database, taskID)
	if task.Status != "doing" {
		t.Errorf("expected Status 'doing', got '%s'", task.Status)
	}
	if task.StartedAt == nil {
		t.Error("expected StartedAt to be set")
	}
}

func TestStartTaskAlreadyStarted(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	taskID, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Task 1",
	})

	service.StartTask(database, taskID)

	// Try to start again
	err := service.StartTask(database, taskID)
	if err == nil {
		t.Error("expected error when starting already started task")
	}
}

func TestCompleteTask(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	taskID, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Task 1",
	})

	service.StartTask(database, taskID)

	err := service.CompleteTask(database, taskID, "Task completed successfully")
	if err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}

	task, _ := service.GetTask(database, taskID)
	if task.Status != "done" {
		t.Errorf("expected Status 'done', got '%s'", task.Status)
	}
	if task.Result != "Task completed successfully" {
		t.Errorf("expected Result 'Task completed successfully', got '%s'", task.Result)
	}
	if task.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

func TestCompleteTaskNotStarted(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	taskID, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Task 1",
	})

	// Try to complete without starting
	err := service.CompleteTask(database, taskID, "Result")
	if err == nil {
		t.Error("expected error when completing non-started task")
	}
}

func TestFailTask(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	taskID, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Task 1",
	})

	service.StartTask(database, taskID)

	err := service.FailTask(database, taskID, "Something went wrong")
	if err != nil {
		t.Fatalf("failed to fail task: %v", err)
	}

	task, _ := service.GetTask(database, taskID)
	if task.Status != "failed" {
		t.Errorf("expected Status 'failed', got '%s'", task.Status)
	}
	if task.Error != "Something went wrong" {
		t.Errorf("expected Error 'Something went wrong', got '%s'", task.Error)
	}
	if task.FailedAt == nil {
		t.Error("expected FailedAt to be set")
	}
}

func TestFailTaskNotStarted(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	taskID, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Task 1",
	})

	// Try to fail without starting
	err := service.FailTask(database, taskID, "Error")
	if err == nil {
		t.Error("expected error when failing non-started task")
	}
}

func TestPopTask(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	// Set context, tech, design for manifest
	service.SetContext(database, map[string]interface{}{"project_name": "Test"})
	service.SetTech(database, map[string]interface{}{"backend": "go"})
	service.SetDesign(database, map[string]interface{}{"architecture": "monolith"})

	// Create tasks
	service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 1"})
	service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 2"})

	result, err := service.PopTask(database)
	if err != nil {
		t.Fatalf("failed to pop task: %v", err)
	}

	if result.Task == nil {
		t.Fatal("expected task, got nil")
	}
	if result.Task.Title != "Task 1" {
		t.Errorf("expected Title 'Task 1', got '%s'", result.Task.Title)
	}
	if result.Task.Status != "doing" {
		t.Errorf("expected Status 'doing', got '%s'", result.Task.Status)
	}

	// Check manifest
	if result.Manifest.Context["project_name"] != "Test" {
		t.Errorf("expected manifest.context.project_name 'Test', got '%v'", result.Manifest.Context["project_name"])
	}
	if result.Manifest.Tech["backend"] != "go" {
		t.Errorf("expected manifest.tech.backend 'go', got '%v'", result.Manifest.Tech["backend"])
	}
}

func TestPopTaskNoPendingTasks(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	result, err := service.PopTask(database)
	if err != nil {
		t.Fatalf("failed to pop task: %v", err)
	}

	if result.Task != nil {
		t.Errorf("expected nil task when no pending tasks, got %v", result.Task)
	}
}

func TestPopTaskStartsFeature(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 1"})

	// Feature should be pending initially
	feature, _ := service.GetFeature(database, featureID)
	if feature.Status != "pending" {
		t.Errorf("expected initial feature status 'pending', got '%s'", feature.Status)
	}

	// Pop task should start the feature
	service.PopTask(database)

	feature, _ = service.GetFeature(database, featureID)
	if feature.Status != "active" {
		t.Errorf("expected feature status 'active' after pop, got '%s'", feature.Status)
	}
}

func TestGetTaskStatus(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	// Create tasks with different statuses
	task1, _ := service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 1"})
	task2, _ := service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 2"})
	task3, _ := service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 3"})
	task4, _ := service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 4"})

	// pending: Task 1
	// doing: Task 2
	service.StartTask(database, task2)
	// done: Task 3
	service.StartTask(database, task3)
	service.CompleteTask(database, task3, "done")
	// failed: Task 4
	service.StartTask(database, task4)
	service.FailTask(database, task4, "error")

	// Leave task1 as pending

	status, err := service.GetTaskStatus(database)
	if err != nil {
		t.Fatalf("failed to get task status: %v", err)
	}

	if status.Total != 4 {
		t.Errorf("expected Total 4, got %d", status.Total)
	}
	if status.Pending != 1 {
		t.Errorf("expected Pending 1, got %d", status.Pending)
	}
	if status.Doing != 1 {
		t.Errorf("expected Doing 1, got %d", status.Doing)
	}
	if status.Done != 1 {
		t.Errorf("expected Done 1, got %d", status.Done)
	}
	if status.Failed != 1 {
		t.Errorf("expected Failed 1, got %d", status.Failed)
	}

	// Progress = done / total = 1/4 = 25%
	if status.Progress != 25.0 {
		t.Errorf("expected Progress 25.0, got %f", status.Progress)
	}

	_ = task1 // just to use the variable
}

func TestGetTaskStatusEmpty(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	status, err := service.GetTaskStatus(database)
	if err != nil {
		t.Fatalf("failed to get task status: %v", err)
	}

	if status.Total != 0 {
		t.Errorf("expected Total 0, got %d", status.Total)
	}
	if status.Progress != 0 {
		t.Errorf("expected Progress 0, got %f", status.Progress)
	}
}

func TestPopTaskWithHighPriorityMemos(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature 1", "Description")

	// Create memo with priority 1
	service.SetMemo(database, service.MemoSetInput{
		Scope:    "project",
		ScopeID:  "test-project",
		Key:      "important",
		Value:    "Important note",
		Priority: 1,
	})

	// Create memo with priority 2 (should not be included)
	service.SetMemo(database, service.MemoSetInput{
		Scope:    "project",
		ScopeID:  "test-project",
		Key:      "normal",
		Value:    "Normal note",
		Priority: 2,
	})

	service.CreateTask(database, service.TaskCreateInput{FeatureID: featureID, Title: "Task 1"})

	result, err := service.PopTask(database)
	if err != nil {
		t.Fatalf("failed to pop task: %v", err)
	}

	// Should only have priority 1 memo
	if len(result.Manifest.Memos) != 1 {
		t.Errorf("expected 1 memo, got %d", len(result.Manifest.Memos))
	}
	if len(result.Manifest.Memos) > 0 && result.Manifest.Memos[0].Key != "important" {
		t.Errorf("expected memo key 'important', got '%s'", result.Manifest.Memos[0].Key)
	}
}
