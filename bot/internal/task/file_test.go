package task

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontmatterRoundTrip(t *testing.T) {
	fm := Frontmatter{Status: "todo"}
	content := FormatFrontmatter(fm, "Test Title", "Body content")

	parsedFM, title, body, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if parsedFM.Status != "todo" {
		t.Errorf("status = %q, want %q", parsedFM.Status, "todo")
	}
	if title != "Test Title" {
		t.Errorf("title = %q, want %q", title, "Test Title")
	}
	if body != "Body content" {
		t.Errorf("body = %q, want %q", body, "Body content")
	}
}

func TestFrontmatterWithParent(t *testing.T) {
	parent := 3
	fm := Frontmatter{Status: "planned", Parent: &parent, Priority: 5}
	content := FormatFrontmatter(fm, "Child Task", "Some spec")

	parsedFM, title, _, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if parsedFM.Status != "planned" {
		t.Errorf("status = %q, want %q", parsedFM.Status, "planned")
	}
	if parsedFM.Parent == nil || *parsedFM.Parent != 3 {
		t.Errorf("parent = %v, want 3", parsedFM.Parent)
	}
	if parsedFM.Priority != 5 {
		t.Errorf("priority = %d, want 5", parsedFM.Priority)
	}
	if title != "Child Task" {
		t.Errorf("title = %q, want %q", title, "Child Task")
	}
}

func TestFrontmatterEmptyBody(t *testing.T) {
	fm := Frontmatter{Status: "todo"}
	content := FormatFrontmatter(fm, "Title Only", "")

	_, title, body, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if title != "Title Only" {
		t.Errorf("title = %q, want %q", title, "Title Only")
	}
	if body != "" {
		t.Errorf("body = %q, want empty", body)
	}
}

func TestParseFrontmatterErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"no delimiter", "just some text"},
		{"no closing delimiter", "---\nstatus: todo\n# Title"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := ParseFrontmatter(tt.content)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestFileRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	fm := Frontmatter{Status: "todo"}
	err := WriteTaskContent(tmpDir, 1, fm, "Test Task", "Task body")
	if err != nil {
		t.Fatalf("WriteTaskContent: %v", err)
	}

	tc, err := ReadTaskContent(tmpDir, 1)
	if err != nil {
		t.Fatalf("ReadTaskContent: %v", err)
	}
	if tc.Title != "Test Task" {
		t.Errorf("title = %q, want %q", tc.Title, "Test Task")
	}
	if tc.Body != "Task body" {
		t.Errorf("body = %q, want %q", tc.Body, "Task body")
	}
	if tc.Frontmatter.Status != "todo" {
		t.Errorf("status = %q, want %q", tc.Frontmatter.Status, "todo")
	}
}

func TestPlanReportRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	// Read non-existent files returns ("", nil)
	plan, err := ReadPlanContent(tmpDir, 1)
	if err != nil {
		t.Fatalf("ReadPlanContent: %v", err)
	}
	if plan != "" {
		t.Errorf("plan = %q, want empty", plan)
	}

	report, err := ReadReportContent(tmpDir, 1)
	if err != nil {
		t.Fatalf("ReadReportContent: %v", err)
	}
	if report != "" {
		t.Errorf("report = %q, want empty", report)
	}

	// Write and read back
	err = WritePlanContent(tmpDir, 1, "# Plan\n\nStep 1\nStep 2")
	if err != nil {
		t.Fatalf("WritePlanContent: %v", err)
	}
	plan, err = ReadPlanContent(tmpDir, 1)
	if err != nil {
		t.Fatalf("ReadPlanContent: %v", err)
	}
	if plan != "# Plan\n\nStep 1\nStep 2" {
		t.Errorf("plan = %q, want plan content", plan)
	}

	err = WriteReportContent(tmpDir, 1, "# Report\n\nDone")
	if err != nil {
		t.Fatalf("WriteReportContent: %v", err)
	}
	report, err = ReadReportContent(tmpDir, 1)
	if err != nil {
		t.Fatalf("ReadReportContent: %v", err)
	}
	if report != "# Report\n\nDone" {
		t.Errorf("report = %q, want report content", report)
	}
}

func TestPathHelpers(t *testing.T) {
	dir := TaskDir("/projects/myapp")
	if !strings.HasSuffix(dir, filepath.Join(".claribot", "tasks")) {
		t.Errorf("TaskDir = %q, want suffix .claribot/tasks", dir)
	}

	path := TaskFilePath("/projects/myapp", 42)
	if !strings.HasSuffix(path, filepath.Join("tasks", "42.md")) {
		t.Errorf("TaskFilePath = %q, want suffix tasks/42.md", path)
	}

	path = PlanFilePath("/projects/myapp", 42)
	if !strings.HasSuffix(path, filepath.Join("tasks", "42.plan.md")) {
		t.Errorf("PlanFilePath = %q, want suffix tasks/42.plan.md", path)
	}

	path = ReportFilePath("/projects/myapp", 42)
	if !strings.HasSuffix(path, filepath.Join("tasks", "42.report.md")) {
		t.Errorf("ReportFilePath = %q, want suffix tasks/42.report.md", path)
	}
}

func TestEnsureTaskDir(t *testing.T) {
	tmpDir := t.TempDir()
	err := EnsureTaskDir(tmpDir)
	if err != nil {
		t.Fatalf("EnsureTaskDir: %v", err)
	}

	info, err := os.Stat(TaskDir(tmpDir))
	if err != nil {
		t.Fatalf("Stat task dir: %v", err)
	}
	if !info.IsDir() {
		t.Error("task dir is not a directory")
	}

	// Calling again should not fail
	err = EnsureTaskDir(tmpDir)
	if err != nil {
		t.Fatalf("EnsureTaskDir second call: %v", err)
	}
}

func TestValidateTaskFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Valid file
	fm := Frontmatter{Status: "todo"}
	err := WriteTaskContent(tmpDir, 1, fm, "Valid Task", "Body content")
	if err != nil {
		t.Fatalf("WriteTaskContent: %v", err)
	}
	err = ValidateTaskFile(TaskFilePath(tmpDir, 1))
	if err != nil {
		t.Errorf("expected valid, got: %v", err)
	}

	// Valid with empty body (warning only, not error)
	err = WriteTaskContent(tmpDir, 2, fm, "No Body", "")
	if err != nil {
		t.Fatalf("WriteTaskContent: %v", err)
	}
	err = ValidateTaskFile(TaskFilePath(tmpDir, 2))
	if err != nil {
		t.Errorf("empty body should be warning only, got error: %v", err)
	}

	// Invalid: bad status
	badFM := Frontmatter{Status: "invalid"}
	err = WriteTaskContent(tmpDir, 3, badFM, "Bad Status", "Body")
	if err != nil {
		t.Fatalf("WriteTaskContent: %v", err)
	}
	err = ValidateTaskFile(TaskFilePath(tmpDir, 3))
	if err == nil {
		t.Error("expected error for invalid status")
	}

	// Invalid: non-existent file
	err = ValidateTaskFile(filepath.Join(tmpDir, "nonexistent.md"))
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Invalid: negative parent
	negParent := -1
	badParentFM := Frontmatter{Status: "todo", Parent: &negParent}
	err = WriteTaskContent(tmpDir, 4, badParentFM, "Bad Parent", "Body")
	if err != nil {
		t.Fatalf("WriteTaskContent: %v", err)
	}
	err = ValidateTaskFile(TaskFilePath(tmpDir, 4))
	if err == nil {
		t.Error("expected error for negative parent")
	}
}
