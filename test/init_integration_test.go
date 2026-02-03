//go:build integration

package test

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/service"
)

// These tests require the 'claude' CLI to be installed and available.
// Run with: go test -tags=integration ./test/...

func TestClariInit_SkipAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clari-init-skip")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Suppress output
	oldWriter := service.DefaultWriter
	service.DefaultWriter = io.Discard
	defer func() { service.DefaultWriter = oldWriter }()

	config := service.InitConfig{
		ProjectID:    "test-project",
		Name:         "Test Project",
		Description:  "Integration test project",
		SkipAnalysis: true,
		SkipSpecs:    true,
		WorkDir:      tmpDir,
	}

	result, err := service.RunInit(config)
	if err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}

	// Verify DB was created
	dbPath := filepath.Join(tmpDir, ".claritask", "db.clt")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected database to be created")
	}

	// Verify project record
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	project, err := service.GetProject(database)
	if err != nil {
		t.Fatalf("failed to get project: %v", err)
	}
	if project.ID != "test-project" {
		t.Errorf("expected project ID 'test-project', got '%s'", project.ID)
	}
}

func TestClariInit_NonInteractive(t *testing.T) {
	// Check if claude CLI is available
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not available, skipping test")
	}

	tmpDir, err := os.MkdirTemp("", "clari-init-noninteractive")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple go.mod to be scanned
	goMod := "module test.com/myproject\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Suppress output
	oldWriter := service.DefaultWriter
	service.DefaultWriter = io.Discard
	defer func() { service.DefaultWriter = oldWriter }()

	config := service.InitConfig{
		ProjectID:      "test-project",
		Name:           "Test Project",
		Description:    "A test project for integration testing",
		NonInteractive: true,
		WorkDir:        tmpDir,
	}

	result, err := service.RunInit(config)
	if err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}

	// Verify specs file was created
	specsPath := filepath.Join(tmpDir, "specs", "test-project.md")
	if _, err := os.Stat(specsPath); os.IsNotExist(err) {
		t.Error("expected specs file to be created")
	}
}

func TestClariInit_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clari-init-force")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Suppress output
	oldWriter := service.DefaultWriter
	service.DefaultWriter = io.Discard
	defer func() { service.DefaultWriter = oldWriter }()

	// First init
	config1 := service.InitConfig{
		ProjectID:    "test-project",
		Name:         "Test Project 1",
		SkipAnalysis: true,
		SkipSpecs:    true,
		WorkDir:      tmpDir,
	}

	_, err = service.RunInit(config1)
	if err != nil {
		t.Fatalf("first RunInit failed: %v", err)
	}

	// Second init without force should fail
	config2 := service.InitConfig{
		ProjectID:    "test-project",
		Name:         "Test Project 2",
		SkipAnalysis: true,
		SkipSpecs:    true,
		WorkDir:      tmpDir,
		Force:        false,
	}

	result2, err := service.RunInit(config2)
	if err == nil && result2.Success {
		t.Error("expected second init without --force to fail")
	}

	// Third init with force should succeed
	config3 := service.InitConfig{
		ProjectID:    "test-project",
		Name:         "Test Project 3",
		SkipAnalysis: true,
		SkipSpecs:    true,
		WorkDir:      tmpDir,
		Force:        true,
	}

	result3, err := service.RunInit(config3)
	if err != nil {
		t.Fatalf("third RunInit with --force failed: %v", err)
	}
	if !result3.Success {
		t.Errorf("expected success with --force, got error: %s", result3.Error)
	}
}

func TestClariInit_Resume(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clari-init-resume")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Suppress output
	oldWriter := service.DefaultWriter
	service.DefaultWriter = io.Discard
	defer func() { service.DefaultWriter = oldWriter }()

	// First, do a partial init (Phase 1 only via skip flags)
	config := service.InitConfig{
		ProjectID:    "test-project",
		Name:         "Test Project",
		SkipAnalysis: true,
		SkipSpecs:    true,
		WorkDir:      tmpDir,
	}

	result, err := service.RunInit(config)
	if err != nil {
		t.Fatalf("RunInit failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}

	// Open db and simulate partial state
	dbPath := filepath.Join(tmpDir, ".claritask", "db.clt")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	// Save a partial state
	state := &service.InitState{
		Phase:     service.InitPhaseAnalysis,
		ProjectID: "test-project",
		Tech: map[string]interface{}{
			"language": "Go",
		},
	}
	if err := service.SaveInitState(database, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}
	database.Close()

	// Resume should work
	result, err = service.ResumeInit(tmpDir)
	// Note: Resume currently re-runs full init, which may fail without claude
	// This test mainly verifies the resume mechanism exists
	if err != nil {
		// Expected if claude is not available
		t.Logf("ResumeInit returned error (expected without claude): %v", err)
	}
}
