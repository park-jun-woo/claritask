package test

import (
	"os"
	"path/filepath"
	"testing"

	"parkjunwoo.com/talos/internal/db"
	"parkjunwoo.com/talos/internal/service"
)

func setupTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "talos-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open database: %v", err)
	}

	if err := database.Migrate(); err != nil {
		database.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to migrate: %v", err)
	}

	cleanup := func() {
		database.Close()
		os.RemoveAll(tmpDir)
	}

	return database, cleanup
}

func TestCreateProject(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	err := service.CreateProject(database, "test-project", "Test Project", "A test project")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Verify project was created
	project, err := service.GetProject(database)
	if err != nil {
		t.Fatalf("failed to get project: %v", err)
	}

	if project.ID != "test-project" {
		t.Errorf("expected ID 'test-project', got '%s'", project.ID)
	}
	if project.Name != "Test Project" {
		t.Errorf("expected Name 'Test Project', got '%s'", project.Name)
	}
	if project.Description != "A test project" {
		t.Errorf("expected Description 'A test project', got '%s'", project.Description)
	}
	if project.Status != "active" {
		t.Errorf("expected Status 'active', got '%s'", project.Status)
	}
}

func TestCreateProjectDuplicateID(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	err := service.CreateProject(database, "test-project", "Test Project", "Description")
	if err != nil {
		t.Fatalf("failed to create first project: %v", err)
	}

	// Try to create with same ID
	err = service.CreateProject(database, "test-project", "Another Project", "Description")
	if err == nil {
		t.Error("expected error when creating project with duplicate ID")
	}
}

func TestGetProjectNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := service.GetProject(database)
	if err == nil {
		t.Error("expected error when getting non-existent project")
	}
}

func TestUpdateProject(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	err := service.CreateProject(database, "test-project", "Test Project", "Description")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	project, err := service.GetProject(database)
	if err != nil {
		t.Fatalf("failed to get project: %v", err)
	}

	project.Name = "Updated Name"
	project.Description = "Updated Description"
	project.Status = "archived"

	err = service.UpdateProject(database, project)
	if err != nil {
		t.Fatalf("failed to update project: %v", err)
	}

	updated, err := service.GetProject(database)
	if err != nil {
		t.Fatalf("failed to get updated project: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("expected Name 'Updated Name', got '%s'", updated.Name)
	}
	if updated.Description != "Updated Description" {
		t.Errorf("expected Description 'Updated Description', got '%s'", updated.Description)
	}
	if updated.Status != "archived" {
		t.Errorf("expected Status 'archived', got '%s'", updated.Status)
	}
}

func TestSetContext(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	data := map[string]interface{}{
		"project_name": "Test Project",
		"description":  "A test project",
	}

	err := service.SetContext(database, data)
	if err != nil {
		t.Fatalf("failed to set context: %v", err)
	}

	// Get context
	result, err := service.GetContext(database)
	if err != nil {
		t.Fatalf("failed to get context: %v", err)
	}

	if result["project_name"] != "Test Project" {
		t.Errorf("expected project_name 'Test Project', got '%v'", result["project_name"])
	}
	if result["description"] != "A test project" {
		t.Errorf("expected description 'A test project', got '%v'", result["description"])
	}
}

func TestSetContextUpsert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// First set
	data1 := map[string]interface{}{"key1": "value1"}
	err := service.SetContext(database, data1)
	if err != nil {
		t.Fatalf("failed to set context: %v", err)
	}

	// Update
	data2 := map[string]interface{}{"key1": "updated", "key2": "value2"}
	err = service.SetContext(database, data2)
	if err != nil {
		t.Fatalf("failed to update context: %v", err)
	}

	result, err := service.GetContext(database)
	if err != nil {
		t.Fatalf("failed to get context: %v", err)
	}

	if result["key1"] != "updated" {
		t.Errorf("expected key1 'updated', got '%v'", result["key1"])
	}
	if result["key2"] != "value2" {
		t.Errorf("expected key2 'value2', got '%v'", result["key2"])
	}
}

func TestGetContextNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := service.GetContext(database)
	if err == nil {
		t.Error("expected error when getting non-existent context")
	}
}

func TestSetTech(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	data := map[string]interface{}{
		"backend":  "go",
		"frontend": "react",
		"database": "postgresql",
	}

	err := service.SetTech(database, data)
	if err != nil {
		t.Fatalf("failed to set tech: %v", err)
	}

	result, err := service.GetTech(database)
	if err != nil {
		t.Fatalf("failed to get tech: %v", err)
	}

	if result["backend"] != "go" {
		t.Errorf("expected backend 'go', got '%v'", result["backend"])
	}
	if result["frontend"] != "react" {
		t.Errorf("expected frontend 'react', got '%v'", result["frontend"])
	}
}

func TestSetTechUpsert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	data1 := map[string]interface{}{"backend": "python"}
	err := service.SetTech(database, data1)
	if err != nil {
		t.Fatalf("failed to set tech: %v", err)
	}

	data2 := map[string]interface{}{"backend": "go", "frontend": "vue"}
	err = service.SetTech(database, data2)
	if err != nil {
		t.Fatalf("failed to update tech: %v", err)
	}

	result, err := service.GetTech(database)
	if err != nil {
		t.Fatalf("failed to get tech: %v", err)
	}

	if result["backend"] != "go" {
		t.Errorf("expected backend 'go', got '%v'", result["backend"])
	}
}

func TestGetTechNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := service.GetTech(database)
	if err == nil {
		t.Error("expected error when getting non-existent tech")
	}
}

func TestSetDesign(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	data := map[string]interface{}{
		"architecture": "monolith",
		"auth_method":  "jwt",
		"api_style":    "rest",
	}

	err := service.SetDesign(database, data)
	if err != nil {
		t.Fatalf("failed to set design: %v", err)
	}

	result, err := service.GetDesign(database)
	if err != nil {
		t.Fatalf("failed to get design: %v", err)
	}

	if result["architecture"] != "monolith" {
		t.Errorf("expected architecture 'monolith', got '%v'", result["architecture"])
	}
	if result["auth_method"] != "jwt" {
		t.Errorf("expected auth_method 'jwt', got '%v'", result["auth_method"])
	}
}

func TestSetDesignUpsert(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	data1 := map[string]interface{}{"architecture": "microservice"}
	err := service.SetDesign(database, data1)
	if err != nil {
		t.Fatalf("failed to set design: %v", err)
	}

	data2 := map[string]interface{}{"architecture": "monolith", "api_style": "graphql"}
	err = service.SetDesign(database, data2)
	if err != nil {
		t.Fatalf("failed to update design: %v", err)
	}

	result, err := service.GetDesign(database)
	if err != nil {
		t.Fatalf("failed to get design: %v", err)
	}

	if result["architecture"] != "monolith" {
		t.Errorf("expected architecture 'monolith', got '%v'", result["architecture"])
	}
}

func TestGetDesignNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := service.GetDesign(database)
	if err == nil {
		t.Error("expected error when getting non-existent design")
	}
}

func TestCheckRequiredAllPresent(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Set all required fields
	service.SetContext(database, map[string]interface{}{
		"project_name": "Test",
		"description":  "Description",
	})
	service.SetTech(database, map[string]interface{}{
		"backend":  "go",
		"frontend": "react",
		"database": "postgresql",
	})
	service.SetDesign(database, map[string]interface{}{
		"architecture": "monolith",
		"auth_method":  "jwt",
		"api_style":    "rest",
	})

	result, err := service.CheckRequired(database)
	if err != nil {
		t.Fatalf("failed to check required: %v", err)
	}

	if !result.Ready {
		t.Error("expected Ready=true when all fields present")
	}
	if len(result.MissingRequired) != 0 {
		t.Errorf("expected no missing fields, got %d", len(result.MissingRequired))
	}
}

func TestCheckRequiredMissingFields(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Set only some fields
	service.SetContext(database, map[string]interface{}{
		"project_name": "Test",
		// missing: description
	})

	result, err := service.CheckRequired(database)
	if err != nil {
		t.Fatalf("failed to check required: %v", err)
	}

	if result.Ready {
		t.Error("expected Ready=false when fields are missing")
	}
	if len(result.MissingRequired) == 0 {
		t.Error("expected some missing fields")
	}

	// Check that description is missing
	foundDescription := false
	for _, field := range result.MissingRequired {
		if field.Field == "context.description" {
			foundDescription = true
			break
		}
	}
	if !foundDescription {
		t.Error("expected context.description in missing fields")
	}
}

func TestCheckRequiredNoData(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	result, err := service.CheckRequired(database)
	if err != nil {
		t.Fatalf("failed to check required: %v", err)
	}

	if result.Ready {
		t.Error("expected Ready=false when no data set")
	}
	// Should have all required fields missing (context: 2, tech: 3, design: 3 = 8 total)
	if len(result.MissingRequired) < 8 {
		t.Errorf("expected at least 8 missing fields, got %d", len(result.MissingRequired))
	}
}

func TestSetProjectFull(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create initial project
	err := service.CreateProject(database, "test", "Initial", "Initial Description")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Set full project data
	input := service.ProjectSetInput{
		Name:        "Updated Project",
		Description: "Updated Description",
		Context: map[string]interface{}{
			"project_name": "Test",
			"description":  "Test description",
		},
		Tech: map[string]interface{}{
			"backend": "go",
		},
		Design: map[string]interface{}{
			"architecture": "monolith",
		},
	}

	err = service.SetProjectFull(database, input)
	if err != nil {
		t.Fatalf("failed to set project full: %v", err)
	}

	// Verify project updated
	project, _ := service.GetProject(database)
	if project.Name != "Updated Project" {
		t.Errorf("expected Name 'Updated Project', got '%s'", project.Name)
	}

	// Verify context
	ctx, _ := service.GetContext(database)
	if ctx["project_name"] != "Test" {
		t.Errorf("expected context.project_name 'Test', got '%v'", ctx["project_name"])
	}

	// Verify tech
	tech, _ := service.GetTech(database)
	if tech["backend"] != "go" {
		t.Errorf("expected tech.backend 'go', got '%v'", tech["backend"])
	}

	// Verify design
	design, _ := service.GetDesign(database)
	if design["architecture"] != "monolith" {
		t.Errorf("expected design.architecture 'monolith', got '%v'", design["architecture"])
	}
}

func TestSetProjectFullPartial(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create initial project
	err := service.CreateProject(database, "test", "Initial", "Initial Description")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Set only name (no context, tech, design)
	input := service.ProjectSetInput{
		Name: "New Name",
	}

	err = service.SetProjectFull(database, input)
	if err != nil {
		t.Fatalf("failed to set project full: %v", err)
	}

	project, _ := service.GetProject(database)
	if project.Name != "New Name" {
		t.Errorf("expected Name 'New Name', got '%s'", project.Name)
	}
	// Description should remain unchanged
	if project.Description != "Initial Description" {
		t.Errorf("expected Description 'Initial Description', got '%s'", project.Description)
	}
}
