package test

import (
	"encoding/json"
	"testing"
	"time"

	"parkjunwoo.com/talos/internal/model"
)

func TestProjectStruct(t *testing.T) {
	t.Helper()
	now := time.Now()
	p := model.Project{
		ID:          "test-project",
		Name:        "Test Project",
		Description: "A test project",
		Status:      "active",
		CreatedAt:   now,
	}

	if p.ID != "test-project" {
		t.Errorf("expected ID 'test-project', got '%s'", p.ID)
	}
	if p.Name != "Test Project" {
		t.Errorf("expected Name 'Test Project', got '%s'", p.Name)
	}
	if p.Description != "A test project" {
		t.Errorf("expected Description 'A test project', got '%s'", p.Description)
	}
	if p.Status != "active" {
		t.Errorf("expected Status 'active', got '%s'", p.Status)
	}
	if !p.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, p.CreatedAt)
	}
}

func TestPhaseStruct(t *testing.T) {
	t.Helper()
	now := time.Now()
	p := model.Phase{
		ID:          "1",
		ProjectID:   "test-project",
		Name:        "Phase 1",
		Description: "First phase",
		OrderNum:    1,
		Status:      "pending",
		CreatedAt:   now,
	}

	if p.ID != "1" {
		t.Errorf("expected ID '1', got '%s'", p.ID)
	}
	if p.ProjectID != "test-project" {
		t.Errorf("expected ProjectID 'test-project', got '%s'", p.ProjectID)
	}
	if p.OrderNum != 1 {
		t.Errorf("expected OrderNum 1, got %d", p.OrderNum)
	}
	if p.Status != "pending" {
		t.Errorf("expected Status 'pending', got '%s'", p.Status)
	}
}

func TestTaskStruct(t *testing.T) {
	t.Helper()
	now := time.Now()
	parentID := "1"
	task := model.Task{
		ID:          "1",
		PhaseID:     "1",
		ParentID:    &parentID,
		Status:      "pending",
		Title:       "Task 1",
		Level:       "leaf",
		Skill:       "coding",
		References:  []string{"ref1", "ref2"},
		Content:     "Task content",
		Result:      "",
		Error:       "",
		CreatedAt:   now,
		StartedAt:   nil,
		CompletedAt: nil,
		FailedAt:    nil,
	}

	if task.ID != "1" {
		t.Errorf("expected ID '1', got '%s'", task.ID)
	}
	if task.ParentID == nil || *task.ParentID != "1" {
		t.Errorf("expected ParentID '1', got '%v'", task.ParentID)
	}
	if len(task.References) != 2 {
		t.Errorf("expected 2 references, got %d", len(task.References))
	}
	if task.StartedAt != nil {
		t.Errorf("expected StartedAt nil, got %v", task.StartedAt)
	}
}

func TestTaskWithNilParentID(t *testing.T) {
	t.Helper()
	task := model.Task{
		ID:       "1",
		PhaseID:  "1",
		ParentID: nil,
		Status:   "pending",
		Title:    "Task 1",
	}

	if task.ParentID != nil {
		t.Errorf("expected ParentID nil, got '%v'", task.ParentID)
	}
}

func TestContextStruct(t *testing.T) {
	t.Helper()
	now := time.Now()
	ctx := model.Context{
		ID:        1,
		Data:      `{"project_name":"test"}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if ctx.ID != 1 {
		t.Errorf("expected ID 1, got %d", ctx.ID)
	}
	if ctx.Data != `{"project_name":"test"}` {
		t.Errorf("expected Data with project_name, got '%s'", ctx.Data)
	}
}

func TestTechStruct(t *testing.T) {
	t.Helper()
	now := time.Now()
	tech := model.Tech{
		ID:        1,
		Data:      `{"backend":"go"}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if tech.ID != 1 {
		t.Errorf("expected ID 1, got %d", tech.ID)
	}
}

func TestDesignStruct(t *testing.T) {
	t.Helper()
	now := time.Now()
	design := model.Design{
		ID:        1,
		Data:      `{"architecture":"monolith"}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if design.ID != 1 {
		t.Errorf("expected ID 1, got %d", design.ID)
	}
}

func TestStateStruct(t *testing.T) {
	t.Helper()
	state := model.State{
		Key:   "current_project",
		Value: "test-project",
	}

	if state.Key != "current_project" {
		t.Errorf("expected Key 'current_project', got '%s'", state.Key)
	}
	if state.Value != "test-project" {
		t.Errorf("expected Value 'test-project', got '%s'", state.Value)
	}
}

func TestMemoStruct(t *testing.T) {
	t.Helper()
	now := time.Now()
	memo := model.Memo{
		Scope:     "project",
		ScopeID:   "test-project",
		Key:       "notes",
		Data:      `{"value":"some notes"}`,
		Priority:  1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if memo.Scope != "project" {
		t.Errorf("expected Scope 'project', got '%s'", memo.Scope)
	}
	if memo.Priority != 1 {
		t.Errorf("expected Priority 1, got %d", memo.Priority)
	}
}

func TestResponseJSON(t *testing.T) {
	t.Helper()
	tests := []struct {
		name     string
		response model.Response
		expected string
	}{
		{
			name:     "success response",
			response: model.Response{Success: true},
			expected: `{"success":true}`,
		},
		{
			name:     "error response",
			response: model.Response{Success: false, Error: "something went wrong"},
			expected: `{"success":false,"error":"something went wrong"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var resp model.Response
			if err := json.Unmarshal(data, &resp); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if resp.Success != tt.response.Success {
				t.Errorf("expected Success %v, got %v", tt.response.Success, resp.Success)
			}
			if resp.Error != tt.response.Error {
				t.Errorf("expected Error '%s', got '%s'", tt.response.Error, resp.Error)
			}
		})
	}
}

func TestMemoDataJSON(t *testing.T) {
	t.Helper()
	memoData := model.MemoData{
		Scope:    "project",
		ScopeID:  "test",
		Key:      "config",
		Data:     map[string]interface{}{"setting": "value"},
		Priority: 1,
	}

	data, err := json.Marshal(memoData)
	if err != nil {
		t.Fatalf("failed to marshal MemoData: %v", err)
	}

	var result model.MemoData
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal MemoData: %v", err)
	}

	if result.Scope != "project" {
		t.Errorf("expected Scope 'project', got '%s'", result.Scope)
	}
	if result.Priority != 1 {
		t.Errorf("expected Priority 1, got %d", result.Priority)
	}
	if result.Data["setting"] != "value" {
		t.Errorf("expected Data['setting'] = 'value', got '%v'", result.Data["setting"])
	}
}

func TestManifestJSON(t *testing.T) {
	t.Helper()
	manifest := model.Manifest{
		Context: map[string]interface{}{"project_name": "test"},
		Tech:    map[string]interface{}{"backend": "go"},
		Design:  map[string]interface{}{"architecture": "monolith"},
		State:   map[string]string{"current_project": "test"},
		Memos: []model.MemoData{
			{Scope: "project", Key: "notes", Priority: 1},
		},
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("failed to marshal Manifest: %v", err)
	}

	var result model.Manifest
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal Manifest: %v", err)
	}

	if result.Context["project_name"] != "test" {
		t.Errorf("expected Context.project_name = 'test', got '%v'", result.Context["project_name"])
	}
	if result.Tech["backend"] != "go" {
		t.Errorf("expected Tech.backend = 'go', got '%v'", result.Tech["backend"])
	}
	if result.State["current_project"] != "test" {
		t.Errorf("expected State.current_project = 'test', got '%v'", result.State["current_project"])
	}
	if len(result.Memos) != 1 {
		t.Errorf("expected 1 memo, got %d", len(result.Memos))
	}
}

func TestTaskPopResponseJSON(t *testing.T) {
	t.Helper()
	now := time.Now()
	task := &model.Task{
		ID:        "1",
		PhaseID:   "1",
		Status:    "doing",
		Title:     "Test Task",
		CreatedAt: now,
	}
	manifest := model.Manifest{
		Context: map[string]interface{}{"project_name": "test"},
		Tech:    map[string]interface{}{},
		Design:  map[string]interface{}{},
		State:   map[string]string{},
		Memos:   []model.MemoData{},
	}

	response := model.TaskPopResponse{
		Task:     task,
		Manifest: manifest,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal TaskPopResponse: %v", err)
	}

	var result model.TaskPopResponse
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal TaskPopResponse: %v", err)
	}

	if result.Task == nil {
		t.Fatal("expected Task not nil")
	}
	if result.Task.Title != "Test Task" {
		t.Errorf("expected Task.Title 'Test Task', got '%s'", result.Task.Title)
	}
}

func TestEmptyReferences(t *testing.T) {
	t.Helper()
	task := model.Task{
		ID:         "1",
		PhaseID:    "1",
		Status:     "pending",
		Title:      "Task",
		References: []string{},
	}

	if task.References == nil {
		t.Error("expected References to be empty slice, not nil")
	}
	if len(task.References) != 0 {
		t.Errorf("expected 0 references, got %d", len(task.References))
	}
}
