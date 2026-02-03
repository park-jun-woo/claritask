package test

import (
	"testing"

	"parkjunwoo.com/claritask/internal/service"
)

func TestSetMemo(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	input := service.MemoSetInput{
		Scope:   "project",
		ScopeID: "test-project",
		Key:     "notes",
		Value:   "Some important notes",
	}

	err := service.SetMemo(database, input)
	if err != nil {
		t.Fatalf("failed to set memo: %v", err)
	}

	// Verify memo was created
	memo, err := service.GetMemo(database, "project", "test-project", "notes")
	if err != nil {
		t.Fatalf("failed to get memo: %v", err)
	}

	if memo.Scope != "project" {
		t.Errorf("expected Scope 'project', got '%s'", memo.Scope)
	}
	if memo.Key != "notes" {
		t.Errorf("expected Key 'notes', got '%s'", memo.Key)
	}
	// Default priority should be 2
	if memo.Priority != 2 {
		t.Errorf("expected Priority 2, got %d", memo.Priority)
	}
}

func TestSetMemoUpsert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// First set
	input1 := service.MemoSetInput{
		Scope:   "project",
		ScopeID: "test-project",
		Key:     "notes",
		Value:   "Original value",
	}
	service.SetMemo(database, input1)

	// Update
	input2 := service.MemoSetInput{
		Scope:    "project",
		ScopeID:  "test-project",
		Key:      "notes",
		Value:    "Updated value",
		Priority: 1,
	}
	err := service.SetMemo(database, input2)
	if err != nil {
		t.Fatalf("failed to update memo: %v", err)
	}

	memo, _ := service.GetMemo(database, "project", "test-project", "notes")
	if memo.Priority != 1 {
		t.Errorf("expected Priority 1 after update, got %d", memo.Priority)
	}
}

func TestSetMemoWithSummaryAndTags(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	input := service.MemoSetInput{
		Scope:    "project",
		ScopeID:  "test-project",
		Key:      "config",
		Value:    "Configuration details",
		Summary:  "Project configuration",
		Tags:     []string{"config", "important"},
		Priority: 1,
	}

	err := service.SetMemo(database, input)
	if err != nil {
		t.Fatalf("failed to set memo: %v", err)
	}

	memo, _ := service.GetMemo(database, "project", "test-project", "config")
	// Data should contain value, summary, and tags
	if memo.Data == "" {
		t.Error("expected Data to be set")
	}
}

func TestGetMemo(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.SetMemo(database, service.MemoSetInput{
		Scope:   "feature",
		ScopeID: "1",
		Key:     "decisions",
		Value:   "Feature decisions",
	})

	memo, err := service.GetMemo(database, "feature", "1", "decisions")
	if err != nil {
		t.Fatalf("failed to get memo: %v", err)
	}

	if memo.Scope != "feature" {
		t.Errorf("expected Scope 'feature', got '%s'", memo.Scope)
	}
	if memo.ScopeID != "1" {
		t.Errorf("expected ScopeID '1', got '%s'", memo.ScopeID)
	}
}

func TestGetMemoNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := service.GetMemo(database, "project", "test", "nonexistent")
	if err == nil {
		t.Error("expected error when getting non-existent memo")
	}
}

func TestDeleteMemo(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.SetMemo(database, service.MemoSetInput{
		Scope:   "project",
		ScopeID: "test",
		Key:     "notes",
		Value:   "To be deleted",
	})

	err := service.DeleteMemo(database, "project", "test", "notes")
	if err != nil {
		t.Fatalf("failed to delete memo: %v", err)
	}

	// Verify memo is deleted
	_, err = service.GetMemo(database, "project", "test", "notes")
	if err == nil {
		t.Error("expected error when getting deleted memo")
	}
}

func TestDeleteMemoNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	err := service.DeleteMemo(database, "project", "test", "nonexistent")
	if err == nil {
		t.Error("expected error when deleting non-existent memo")
	}
}

func TestListMemos(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create memos in different scopes
	service.SetMemo(database, service.MemoSetInput{
		Scope:   "project",
		ScopeID: "proj1",
		Key:     "notes1",
		Value:   "Project notes",
	})
	service.SetMemo(database, service.MemoSetInput{
		Scope:   "feature",
		ScopeID: "1",
		Key:     "notes2",
		Value:   "Feature notes",
	})
	service.SetMemo(database, service.MemoSetInput{
		Scope:   "task",
		ScopeID: "1:1",
		Key:     "notes3",
		Value:   "Task notes",
	})

	result, err := service.ListMemos(database)
	if err != nil {
		t.Fatalf("failed to list memos: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected Total 3, got %d", result.Total)
	}
	if len(result.Project) == 0 {
		t.Error("expected project memos")
	}
	if len(result.Feature) == 0 {
		t.Error("expected feature memos")
	}
	if len(result.Task) == 0 {
		t.Error("expected task memos")
	}
}

func TestListMemosByScopeCategorization(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create multiple memos in same scope
	service.SetMemo(database, service.MemoSetInput{
		Scope:   "project",
		ScopeID: "proj1",
		Key:     "key1",
		Value:   "Value 1",
		Summary: "Summary 1",
	})
	service.SetMemo(database, service.MemoSetInput{
		Scope:   "project",
		ScopeID: "proj1",
		Key:     "key2",
		Value:   "Value 2",
	})

	result, _ := service.ListMemos(database)

	if len(result.Project["proj1"]) != 2 {
		t.Errorf("expected 2 memos for proj1, got %d", len(result.Project["proj1"]))
	}
}

func TestListMemosByScope(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service.SetMemo(database, service.MemoSetInput{
		Scope:   "feature",
		ScopeID: "1",
		Key:     "notes1",
		Value:   "Notes 1",
	})
	service.SetMemo(database, service.MemoSetInput{
		Scope:   "feature",
		ScopeID: "1",
		Key:     "notes2",
		Value:   "Notes 2",
	})
	service.SetMemo(database, service.MemoSetInput{
		Scope:   "feature",
		ScopeID: "2",
		Key:     "notes3",
		Value:   "Notes 3",
	})

	memos, err := service.ListMemosByScope(database, "feature", "1")
	if err != nil {
		t.Fatalf("failed to list memos by scope: %v", err)
	}

	if len(memos) != 2 {
		t.Errorf("expected 2 memos, got %d", len(memos))
	}
}

func TestGetHighPriorityMemos(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create memos with different priorities
	service.SetMemo(database, service.MemoSetInput{
		Scope:    "project",
		ScopeID:  "test",
		Key:      "high1",
		Value:    "High priority 1",
		Priority: 1,
	})
	service.SetMemo(database, service.MemoSetInput{
		Scope:    "project",
		ScopeID:  "test",
		Key:      "normal",
		Value:    "Normal priority",
		Priority: 2,
	})
	service.SetMemo(database, service.MemoSetInput{
		Scope:    "project",
		ScopeID:  "test",
		Key:      "high2",
		Value:    "High priority 2",
		Priority: 1,
	})
	service.SetMemo(database, service.MemoSetInput{
		Scope:    "project",
		ScopeID:  "test",
		Key:      "low",
		Value:    "Low priority",
		Priority: 3,
	})

	memos, err := service.GetHighPriorityMemos(database)
	if err != nil {
		t.Fatalf("failed to get high priority memos: %v", err)
	}

	if len(memos) != 2 {
		t.Errorf("expected 2 high priority memos, got %d", len(memos))
	}

	// All should be priority 1
	for _, m := range memos {
		if m.Priority != 1 {
			t.Errorf("expected Priority 1, got %d", m.Priority)
		}
	}
}

func TestGetHighPriorityMemosEmpty(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create only normal priority memos
	service.SetMemo(database, service.MemoSetInput{
		Scope:    "project",
		ScopeID:  "test",
		Key:      "normal",
		Value:    "Normal priority",
		Priority: 2,
	})

	memos, err := service.GetHighPriorityMemos(database)
	if err != nil {
		t.Fatalf("failed to get high priority memos: %v", err)
	}

	if len(memos) != 0 {
		t.Errorf("expected 0 high priority memos, got %d", len(memos))
	}
}

func TestDefaultPriority(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create memo without specifying priority
	input := service.MemoSetInput{
		Scope:   "project",
		ScopeID: "test",
		Key:     "notes",
		Value:   "Notes",
		// Priority not specified, should default to 2
	}
	service.SetMemo(database, input)

	memo, _ := service.GetMemo(database, "project", "test", "notes")
	if memo.Priority != 2 {
		t.Errorf("expected default Priority 2, got %d", memo.Priority)
	}
}

func TestParseMemoKeyProjectLevel(t *testing.T) {
	scope, scopeID, key, err := service.ParseMemoKey("jwt_config")
	if err != nil {
		t.Fatalf("failed to parse memo key: %v", err)
	}

	if scope != "project" {
		t.Errorf("expected scope 'project', got '%s'", scope)
	}
	if scopeID != "" {
		t.Errorf("expected scopeID '', got '%s'", scopeID)
	}
	if key != "jwt_config" {
		t.Errorf("expected key 'jwt_config', got '%s'", key)
	}
}

func TestParseMemoKeyFeatureLevel(t *testing.T) {
	scope, scopeID, key, err := service.ParseMemoKey("PH001:api_decisions")
	if err != nil {
		t.Fatalf("failed to parse memo key: %v", err)
	}

	if scope != "feature" {
		t.Errorf("expected scope 'feature', got '%s'", scope)
	}
	if scopeID != "PH001" {
		t.Errorf("expected scopeID 'PH001', got '%s'", scopeID)
	}
	if key != "api_decisions" {
		t.Errorf("expected key 'api_decisions', got '%s'", key)
	}
}

func TestParseMemoKeyFeatureLevelNumeric(t *testing.T) {
	scope, scopeID, key, err := service.ParseMemoKey("1:notes")
	if err != nil {
		t.Fatalf("failed to parse memo key: %v", err)
	}

	if scope != "feature" {
		t.Errorf("expected scope 'feature', got '%s'", scope)
	}
	if scopeID != "1" {
		t.Errorf("expected scopeID '1', got '%s'", scopeID)
	}
	if key != "notes" {
		t.Errorf("expected key 'notes', got '%s'", key)
	}
}

func TestParseMemoKeyTaskLevel(t *testing.T) {
	scope, scopeID, key, err := service.ParseMemoKey("PH001:T042:notes")
	if err != nil {
		t.Fatalf("failed to parse memo key: %v", err)
	}

	if scope != "task" {
		t.Errorf("expected scope 'task', got '%s'", scope)
	}
	if scopeID != "PH001:T042" {
		t.Errorf("expected scopeID 'PH001:T042', got '%s'", scopeID)
	}
	if key != "notes" {
		t.Errorf("expected key 'notes', got '%s'", key)
	}
}

func TestParseMemoKeyTaskLevelNumeric(t *testing.T) {
	scope, scopeID, key, err := service.ParseMemoKey("1:2:config")
	if err != nil {
		t.Fatalf("failed to parse memo key: %v", err)
	}

	if scope != "task" {
		t.Errorf("expected scope 'task', got '%s'", scope)
	}
	if scopeID != "1:2" {
		t.Errorf("expected scopeID '1:2', got '%s'", scopeID)
	}
	if key != "config" {
		t.Errorf("expected key 'config', got '%s'", key)
	}
}

func TestParseMemoKeyInvalidFormat(t *testing.T) {
	_, _, _, err := service.ParseMemoKey("a:b:c:d")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestParseMemoKeyTable(t *testing.T) {
	tests := []struct {
		input   string
		scope   string
		scopeID string
		key     string
		wantErr bool
	}{
		{"simple_key", "project", "", "simple_key", false},
		{"PH1:feature_key", "feature", "PH1", "feature_key", false},
		{"1:2:task_key", "task", "1:2", "task_key", false},
		{"a:b:c:d", "", "", "", true},
		{"a:b:c:d:e", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			scope, scopeID, key, err := service.ParseMemoKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMemoKey(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if scope != tt.scope {
					t.Errorf("scope = %s, want %s", scope, tt.scope)
				}
				if scopeID != tt.scopeID {
					t.Errorf("scopeID = %s, want %s", scopeID, tt.scopeID)
				}
				if key != tt.key {
					t.Errorf("key = %s, want %s", key, tt.key)
				}
			}
		})
	}
}
