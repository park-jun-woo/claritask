package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"parkjunwoo.com/claritask/internal/service"
)

func TestScanProjectFiles_GoProject(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "scan-test-go")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goMod := `module example.com/myproject

go 1.21

require github.com/spf13/cobra v1.8.0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create README.md
	readme := `# My Project

This is a test project.
`
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(readme), 0644); err != nil {
		t.Fatalf("failed to create README.md: %v", err)
	}

	// Scan
	result, err := service.ScanProjectFiles(tmpDir)
	if err != nil {
		t.Fatalf("failed to scan: %v", err)
	}

	// Verify go.mod
	foundGoMod := false
	foundReadme := false
	for _, f := range result.Files {
		if f.Type == "go_mod" {
			foundGoMod = true
			if !strings.Contains(f.Content, "example.com/myproject") {
				t.Errorf("go.mod content should contain module path")
			}
		}
		if f.Type == "readme" {
			foundReadme = true
		}
	}

	if !foundGoMod {
		t.Error("expected go.mod to be found")
	}
	if !foundReadme {
		t.Error("expected README.md to be found")
	}
}

func TestScanProjectFiles_NodeProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scan-test-node")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{
  "name": "my-node-project",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.0"
  }
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	result, err := service.ScanProjectFiles(tmpDir)
	if err != nil {
		t.Fatalf("failed to scan: %v", err)
	}

	foundPackageJSON := false
	for _, f := range result.Files {
		if f.Type == "package_json" {
			foundPackageJSON = true
			if !strings.Contains(f.Content, "my-node-project") {
				t.Errorf("package.json content should contain project name")
			}
		}
	}

	if !foundPackageJSON {
		t.Error("expected package.json to be found")
	}
}

func TestScanProjectFiles_PythonProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scan-test-python")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create requirements.txt
	requirements := `flask==2.0.0
requests==2.28.0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte(requirements), 0644); err != nil {
		t.Fatalf("failed to create requirements.txt: %v", err)
	}

	// Create pyproject.toml
	pyproject := `[build-system]
requires = ["setuptools", "wheel"]
`
	if err := os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644); err != nil {
		t.Fatalf("failed to create pyproject.toml: %v", err)
	}

	result, err := service.ScanProjectFiles(tmpDir)
	if err != nil {
		t.Fatalf("failed to scan: %v", err)
	}

	foundRequirements := false
	foundPyproject := false
	for _, f := range result.Files {
		if f.Type == "requirements_txt" {
			foundRequirements = true
		}
		if f.Type == "pyproject_toml" {
			foundPyproject = true
		}
	}

	if !foundRequirements {
		t.Error("expected requirements.txt to be found")
	}
	if !foundPyproject {
		t.Error("expected pyproject.toml to be found")
	}
}

func TestScanProjectFiles_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scan-test-empty")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result, err := service.ScanProjectFiles(tmpDir)
	if err != nil {
		t.Fatalf("failed to scan: %v", err)
	}

	if len(result.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(result.Files))
	}
	if len(result.Directories) != 0 {
		t.Errorf("expected 0 directories, got %d", len(result.Directories))
	}
}

func TestScanProjectFiles_DirectoryStructure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scan-test-dirs")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create known directories
	dirs := []string{"src", "internal", "lib", "cmd", "test"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, d), 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}

	result, err := service.ScanProjectFiles(tmpDir)
	if err != nil {
		t.Fatalf("failed to scan: %v", err)
	}

	// Check directories
	foundDirs := make(map[string]bool)
	for _, d := range result.Directories {
		foundDirs[d] = true
	}

	for _, d := range dirs {
		if !foundDirs[d] {
			t.Errorf("expected directory %s to be found", d)
		}
	}
}

func TestScanProjectFiles_LargeFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scan-test-large")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a large README.md with 1000 lines
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("This is line number ")
		sb.WriteString(string(rune('0' + i%10)))
		sb.WriteString("\n")
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(sb.String()), 0644); err != nil {
		t.Fatalf("failed to create README.md: %v", err)
	}

	result, err := service.ScanProjectFiles(tmpDir)
	if err != nil {
		t.Fatalf("failed to scan: %v", err)
	}

	// Find the readme
	for _, f := range result.Files {
		if f.Type == "readme" {
			// Content should be truncated (only first 500 lines)
			lines := strings.Split(f.Content, "\n")
			if len(lines) > 500 {
				t.Errorf("expected at most 500 lines, got %d", len(lines))
			}
		}
	}
}

func TestFormatScanResultForLLM(t *testing.T) {
	result := &service.ScanResult{
		Files: []service.ScannedFile{
			{
				Path:    "go.mod",
				Type:    "go_mod",
				Content: "module example.com/test",
			},
			{
				Path:    "README.md",
				Type:    "readme",
				Content: "# Test Project",
			},
		},
		Directories: []string{"src", "internal"},
	}

	formatted := service.FormatScanResultForLLM(result)

	// Check that it contains the expected sections
	if !strings.Contains(formatted, "## Project Structure") {
		t.Error("formatted output should contain Project Structure header")
	}
	if !strings.Contains(formatted, "### Directories") {
		t.Error("formatted output should contain Directories section")
	}
	if !strings.Contains(formatted, "- src/") {
		t.Error("formatted output should list src directory")
	}
	if !strings.Contains(formatted, "### Configuration Files") {
		t.Error("formatted output should contain Configuration Files section")
	}
	if !strings.Contains(formatted, "go.mod") {
		t.Error("formatted output should contain go.mod")
	}
}
