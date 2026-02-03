package test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/service"
)

func TestInitPhase1_DBInit(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "init-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := service.InitConfig{
		ProjectID:   "test-project",
		Name:        "Test Project",
		Description: "A test project",
		WorkDir:     tmpDir,
	}

	database, err := service.InitPhase1_DBInit(config)
	if err != nil {
		t.Fatalf("InitPhase1_DBInit failed: %v", err)
	}
	defer database.Close()

	// Verify .claritask/db.clt was created
	dbPath := filepath.Join(tmpDir, ".claritask", "db.clt")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected .claritask/db.clt to be created")
	}

	// Verify project was created
	project, err := service.GetProject(database)
	if err != nil {
		t.Fatalf("failed to get project: %v", err)
	}
	if project.ID != "test-project" {
		t.Errorf("expected project ID 'test-project', got '%s'", project.ID)
	}
	if project.Name != "Test Project" {
		t.Errorf("expected project name 'Test Project', got '%s'", project.Name)
	}
}

func TestInitPhase1_DBInit_AlreadyExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "init-test-exists")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create the db file first
	claritaskDir := filepath.Join(tmpDir, ".claritask")
	if err := os.MkdirAll(claritaskDir, 0755); err != nil {
		t.Fatalf("failed to create .claritask dir: %v", err)
	}
	dbPath := filepath.Join(claritaskDir, "db.clt")
	if _, err := os.Create(dbPath); err != nil {
		t.Fatalf("failed to create db file: %v", err)
	}

	// Without --force, should fail
	config := service.InitConfig{
		ProjectID: "test-project",
		Name:      "Test Project",
		WorkDir:   tmpDir,
		Force:     false,
	}

	_, err = service.InitPhase1_DBInit(config)
	if err == nil {
		t.Error("expected error when DB already exists without --force")
	}

	// With --force, should succeed
	config.Force = true
	database, err := service.InitPhase1_DBInit(config)
	if err != nil {
		t.Fatalf("InitPhase1_DBInit with --force failed: %v", err)
	}
	database.Close()
}

func TestSaveInitState_LoadInitState(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Save state
	state := &service.InitState{
		Phase:     2,
		ProjectID: "test-project",
		Tech: map[string]interface{}{
			"language": "Go",
		},
		Design: map[string]interface{}{
			"architecture": "MVC",
		},
		Context: map[string]interface{}{
			"project_type": "CLI",
		},
		SpecsRevision: 1,
	}

	err := service.SaveInitState(database, state)
	if err != nil {
		t.Fatalf("SaveInitState failed: %v", err)
	}

	// Load state
	loaded, err := service.LoadInitState(database)
	if err != nil {
		t.Fatalf("LoadInitState failed: %v", err)
	}

	// Verify
	if loaded.Phase != 2 {
		t.Errorf("expected phase 2, got %d", loaded.Phase)
	}
	if loaded.ProjectID != "test-project" {
		t.Errorf("expected project ID 'test-project', got '%s'", loaded.ProjectID)
	}
	if loaded.Tech["language"] != "Go" {
		t.Errorf("expected tech.language 'Go', got '%v'", loaded.Tech["language"])
	}
	if loaded.SpecsRevision != 1 {
		t.Errorf("expected specs revision 1, got %d", loaded.SpecsRevision)
	}
}

func TestInitConfig_Validation(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		wantErr   bool
	}{
		{"valid lowercase", "my-project", false},
		{"valid with numbers", "project123", false},
		{"valid with underscore", "my_project", false},
		{"invalid uppercase", "MyProject", true},
		{"invalid spaces", "my project", true},
		{"invalid special chars", "my@project", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateProjectID(tt.projectID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectID(%q) error = %v, wantErr %v", tt.projectID, err, tt.wantErr)
			}
		})
	}
}

func TestInitPhase3_Approval_NonInteractive(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a project first
	service.CreateProject(database, "test-project", "Test", "")

	// Save init state
	state := &service.InitState{
		Phase:     1,
		ProjectID: "test-project",
	}
	service.SaveInitState(database, state)

	// Prepare analysis result
	result := &service.ContextAnalysisResult{
		Tech: map[string]interface{}{
			"language": "Go",
		},
		Design: map[string]interface{}{
			"architecture": "Layered",
		},
		Context: map[string]interface{}{
			"project_type": "CLI",
		},
	}

	// Non-interactive should auto-approve
	err := service.InitPhase3_Approval(database, result, true)
	if err != nil {
		t.Fatalf("InitPhase3_Approval (non-interactive) failed: %v", err)
	}

	// Verify data was saved
	tech, err := service.GetTech(database)
	if err != nil {
		t.Fatalf("GetTech failed: %v", err)
	}
	if tech["language"] != "Go" {
		t.Errorf("expected tech.language 'Go', got '%v'", tech["language"])
	}
}

func TestInitPhase5_Feedback_NonInteractive(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Non-interactive should auto-approve
	specs := "# Test Specs\n\n## Features\n- Feature 1"
	result, err := service.InitPhase5_Feedback(database, specs, true)
	if err != nil {
		t.Fatalf("InitPhase5_Feedback (non-interactive) failed: %v", err)
	}

	if result != specs {
		t.Error("non-interactive mode should return specs unchanged")
	}
}

func TestInitPhaseConstants(t *testing.T) {
	if service.InitPhaseDBInit != 1 {
		t.Errorf("expected InitPhaseDBInit=1, got %d", service.InitPhaseDBInit)
	}
	if service.InitPhaseAnalysis != 2 {
		t.Errorf("expected InitPhaseAnalysis=2, got %d", service.InitPhaseAnalysis)
	}
	if service.InitPhaseApproval != 3 {
		t.Errorf("expected InitPhaseApproval=3, got %d", service.InitPhaseApproval)
	}
	if service.InitPhaseSpecsGen != 4 {
		t.Errorf("expected InitPhaseSpecsGen=4, got %d", service.InitPhaseSpecsGen)
	}
	if service.InitPhaseFeedback != 5 {
		t.Errorf("expected InitPhaseFeedback=5, got %d", service.InitPhaseFeedback)
	}
	if service.InitPhaseComplete != 6 {
		t.Errorf("expected InitPhaseComplete=6, got %d", service.InitPhaseComplete)
	}
}

func TestRunInit_SkipAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "init-skip-all")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := service.InitConfig{
		ProjectID:    "skip-test",
		Name:         "Skip Test",
		Description:  "Test skip all",
		SkipAnalysis: true,
		SkipSpecs:    true,
		WorkDir:      tmpDir,
	}

	// Suppress output during test
	oldWriter := service.DefaultWriter
	service.DefaultWriter = io.Discard
	defer func() { service.DefaultWriter = oldWriter }()

	result, err := service.RunInit(config)
	if err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.ProjectID != "skip-test" {
		t.Errorf("expected project ID 'skip-test', got '%s'", result.ProjectID)
	}

	// Verify DB was created
	dbPath := filepath.Join(tmpDir, ".claritask", "db.clt")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected database to be created")
	}
}

func TestRunInit_InvalidProjectID(t *testing.T) {
	config := service.InitConfig{
		ProjectID: "Invalid Project",
		WorkDir:   "/tmp",
	}

	result, err := service.RunInit(config)
	if err == nil {
		t.Error("expected error for invalid project ID")
	}
	if result.Success {
		t.Error("expected failure for invalid project ID")
	}
}

// Helper function for creating a test database
func createTestDB(t *testing.T, dir string) *db.DB {
	dbPath := filepath.Join(dir, ".claritask", "db.clt")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	return database
}
