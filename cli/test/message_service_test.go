package test

import (
	"testing"

	"parkjunwoo.com/claritask/internal/service"
)

func TestCreateMessage(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a project first
	service.CreateProject(database, "test-project", "Test Project", "Test Description")

	// Create message without feature
	message, err := service.CreateMessage(database, "test-project", nil, "Add login feature")
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	if message.ID == 0 {
		t.Error("expected message ID to be set")
	}
	if message.ProjectID != "test-project" {
		t.Errorf("expected ProjectID 'test-project', got '%s'", message.ProjectID)
	}
	if message.Content != "Add login feature" {
		t.Errorf("expected Content 'Add login feature', got '%s'", message.Content)
	}
	if message.Status != "pending" {
		t.Errorf("expected Status 'pending', got '%s'", message.Status)
	}
}

func TestCreateMessageWithFeature(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create project and feature
	service.CreateProject(database, "test-project", "Test Project", "Test Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Login", "Login feature")

	message, err := service.CreateMessage(database, "test-project", &featureID, "Add social login")
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	if message.FeatureID == nil || *message.FeatureID != featureID {
		t.Errorf("expected FeatureID %d, got %v", featureID, message.FeatureID)
	}
}

func TestListMessages(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Test Description")

	// Create multiple messages
	service.CreateMessage(database, "test-project", nil, "Message 1")
	service.CreateMessage(database, "test-project", nil, "Message 2")
	service.CreateMessage(database, "test-project", nil, "Message 3")

	messages, total, err := service.ListMessages(database, "test-project", "", nil, 10)
	if err != nil {
		t.Fatalf("failed to list messages: %v", err)
	}

	if len(messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(messages))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
}

func TestListMessagesWithStatusFilter(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Test Description")

	msg1, _ := service.CreateMessage(database, "test-project", nil, "Message 1")
	service.CreateMessage(database, "test-project", nil, "Message 2")

	// Update first message status
	service.UpdateMessageStatus(database, msg1.ID, "completed", "Done", "")

	messages, _, err := service.ListMessages(database, "test-project", "pending", nil, 10)
	if err != nil {
		t.Fatalf("failed to list messages: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("expected 1 pending message, got %d", len(messages))
	}
}

func TestListMessagesWithLimit(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Test Description")

	for i := 0; i < 5; i++ {
		service.CreateMessage(database, "test-project", nil, "Message")
	}

	messages, total, err := service.ListMessages(database, "test-project", "", nil, 3)
	if err != nil {
		t.Fatalf("failed to list messages: %v", err)
	}

	if len(messages) != 3 {
		t.Errorf("expected 3 messages with limit, got %d", len(messages))
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
}

func TestUpdateMessageStatus(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Test Description")
	message, _ := service.CreateMessage(database, "test-project", nil, "Test message")

	// Update to processing
	err := service.UpdateMessageStatus(database, message.ID, "processing", "", "")
	if err != nil {
		t.Fatalf("failed to update status to processing: %v", err)
	}

	// Update to completed with response
	err = service.UpdateMessageStatus(database, message.ID, "completed", "Analysis complete", "")
	if err != nil {
		t.Fatalf("failed to update status to completed: %v", err)
	}

	// Verify
	detail, _ := service.GetMessage(database, message.ID)
	if detail.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", detail.Status)
	}
	if detail.Response != "Analysis complete" {
		t.Errorf("expected response 'Analysis complete', got '%s'", detail.Response)
	}
	if detail.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestUpdateMessageStatusFailed(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Test Description")
	message, _ := service.CreateMessage(database, "test-project", nil, "Test message")

	err := service.UpdateMessageStatus(database, message.ID, "failed", "", "Connection error")
	if err != nil {
		t.Fatalf("failed to update status to failed: %v", err)
	}

	detail, _ := service.GetMessage(database, message.ID)
	if detail.Status != "failed" {
		t.Errorf("expected status 'failed', got '%s'", detail.Status)
	}
	if detail.Error != "Connection error" {
		t.Errorf("expected error 'Connection error', got '%s'", detail.Error)
	}
}

func TestDeleteMessage(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Test Description")
	message, _ := service.CreateMessage(database, "test-project", nil, "To be deleted")

	err := service.DeleteMessage(database, message.ID)
	if err != nil {
		t.Fatalf("failed to delete message: %v", err)
	}

	// Verify deletion
	_, err = service.GetMessage(database, message.ID)
	if err == nil {
		t.Error("expected error when getting deleted message")
	}
}

func TestDeleteMessageNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	err := service.DeleteMessage(database, 9999)
	if err == nil {
		t.Error("expected error when deleting non-existent message")
	}
}

func TestLinkMessageTask(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Test Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature1", "Description")

	message, _ := service.CreateMessage(database, "test-project", nil, "Add feature")
	taskID, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Implement feature",
		Content:   "Content",
	})

	err := service.LinkMessageTask(database, message.ID, taskID)
	if err != nil {
		t.Fatalf("failed to link message task: %v", err)
	}

	// Verify link via GetMessage
	detail, _ := service.GetMessage(database, message.ID)
	if len(detail.Tasks) != 1 {
		t.Errorf("expected 1 linked task, got %d", len(detail.Tasks))
	}
	if detail.Tasks[0].ID != taskID {
		t.Errorf("expected task ID %d, got %d", taskID, detail.Tasks[0].ID)
	}
}

func TestCountMessageTasks(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Test Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature1", "Description")

	message, _ := service.CreateMessage(database, "test-project", nil, "Add features")
	taskID1, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Task 1",
		Content:   "Content",
	})
	taskID2, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Task 2",
		Content:   "Content",
	})

	service.LinkMessageTask(database, message.ID, taskID1)
	service.LinkMessageTask(database, message.ID, taskID2)

	count, err := service.CountMessageTasks(database, message.ID)
	if err != nil {
		t.Fatalf("failed to count message tasks: %v", err)
	}

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestMessageListItemTasksCount(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.CreateProject(database, "test-project", "Test Project", "Test Description")
	featureID, _ := service.CreateFeature(database, "test-project", "Feature1", "Description")

	message, _ := service.CreateMessage(database, "test-project", nil, "Add features")
	taskID1, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Task 1",
		Content:   "Content",
	})
	taskID2, _ := service.CreateTask(database, service.TaskCreateInput{
		FeatureID: featureID,
		Title:     "Task 2",
		Content:   "Content",
	})

	service.LinkMessageTask(database, message.ID, taskID1)
	service.LinkMessageTask(database, message.ID, taskID2)

	messages, _, _ := service.ListMessages(database, "test-project", "", nil, 10)
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	if messages[0].TasksCount != 2 {
		t.Errorf("expected tasks_count 2, got %d", messages[0].TasksCount)
	}
}
