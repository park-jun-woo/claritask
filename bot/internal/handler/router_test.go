package handler

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple words",
			input: "project list all",
			want:  []string{"project", "list", "all"},
		},
		{
			name:  "double quoted string",
			input: `task add "my task title"`,
			want:  []string{"task", "add", "my task title"},
		},
		{
			name:  "single quoted string",
			input: `task add 'my task title'`,
			want:  []string{"task", "add", "my task title"},
		},
		{
			name:  "mixed quotes",
			input: `schedule add "0 9 * * *" 'run backup' --project test`,
			want:  []string{"schedule", "add", "0 9 * * *", "run backup", "--project", "test"},
		},
		{
			name:  "multiple spaces",
			input: "task   list   all",
			want:  []string{"task", "list", "all"},
		},
		{
			name:  "tabs as separators",
			input: "task\tlist\tall",
			want:  []string{"task", "list", "all"},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "only spaces",
			input: "   ",
			want:  nil,
		},
		{
			name:  "quoted with spaces",
			input: `message send "Hello, this is a test message"`,
			want:  []string{"message", "send", "Hello, this is a test message"},
		},
		{
			name:  "nested quotes preserved",
			input: `task add "Task with 'single' quotes inside"`,
			want:  []string{"task", "add", "Task with 'single' quotes inside"},
		},
		{
			name:  "leading and trailing spaces",
			input: "  project list  ",
			want:  []string{"project", "list"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseArgs(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseArgs(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRouterParsePagination(t *testing.T) {
	r := NewRouter()

	tests := []struct {
		name         string
		args         []string
		wantPage     int
		wantPageSize int
	}{
		{
			name:         "no pagination",
			args:         []string{},
			wantPage:     1,
			wantPageSize: r.pageSize,
		},
		{
			name:         "page only",
			args:         []string{"-p", "3"},
			wantPage:     3,
			wantPageSize: r.pageSize,
		},
		{
			name:         "pageSize only",
			args:         []string{"-n", "20"},
			wantPage:     1,
			wantPageSize: 20,
		},
		{
			name:         "both page and pageSize",
			args:         []string{"-p", "2", "-n", "15"},
			wantPage:     2,
			wantPageSize: 15,
		},
		{
			name:         "mixed with other args",
			args:         []string{"some", "value", "-p", "5", "-n", "10", "other"},
			wantPage:     5,
			wantPageSize: 10,
		},
		{
			name:         "invalid page value",
			args:         []string{"-p", "invalid"},
			wantPage:     1,
			wantPageSize: r.pageSize,
		},
		{
			name:         "zero page",
			args:         []string{"-p", "0"},
			wantPage:     1,
			wantPageSize: r.pageSize,
		},
		{
			name:         "negative page",
			args:         []string{"-p", "-1"},
			wantPage:     1,
			wantPageSize: r.pageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, pageSize := r.parsePagination(tt.args)
			if page != tt.wantPage {
				t.Errorf("page = %d, want %d", page, tt.wantPage)
			}
			if pageSize != tt.wantPageSize {
				t.Errorf("pageSize = %d, want %d", pageSize, tt.wantPageSize)
			}
		})
	}
}

func TestRouterSetPageSize(t *testing.T) {
	r := NewRouter()

	// Check default
	if r.pageSize != 10 {
		t.Errorf("default pageSize = %d, want 10", r.pageSize)
	}

	// Set valid size
	r.SetPageSize(25)
	if r.pageSize != 25 {
		t.Errorf("pageSize after SetPageSize(25) = %d, want 25", r.pageSize)
	}

	// Set invalid size (should be ignored)
	r.SetPageSize(0)
	if r.pageSize != 25 {
		t.Errorf("pageSize after SetPageSize(0) = %d, want 25 (unchanged)", r.pageSize)
	}

	r.SetPageSize(-5)
	if r.pageSize != 25 {
		t.Errorf("pageSize after SetPageSize(-5) = %d, want 25 (unchanged)", r.pageSize)
	}
}

func TestRouterSetProject(t *testing.T) {
	r := NewRouter()

	r.SetProject("test-project", "/path/to/project", "Test Description")

	id, path := r.GetProject()
	if id != "test-project" {
		t.Errorf("ProjectID = %q, want %q", id, "test-project")
	}
	if path != "/path/to/project" {
		t.Errorf("ProjectPath = %q, want %q", path, "/path/to/project")
	}
	if r.ctx.ProjectDescription != "Test Description" {
		t.Errorf("ProjectDescription = %q, want %q", r.ctx.ProjectDescription, "Test Description")
	}
}
