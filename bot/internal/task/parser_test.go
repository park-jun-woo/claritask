package task

import (
	"testing"
)

func TestParsePlanOutput_PlannedDirect(t *testing.T) {
	output := `[PLANNED]
## 구현 방향
간단한 버그 수정

## 변경 파일
- ` + "`path/file.go`" + ` - 수정`
	result := ParsePlanOutput(output)
	if result.Type != "planned" {
		t.Errorf("expected type 'planned', got '%s'", result.Type)
	}
	if result.Plan == "" {
		t.Error("expected non-empty plan")
	}
}

func TestParsePlanOutput_SplitDirect(t *testing.T) {
	output := `[SPLIT]
- Task #10: 첫 번째 하위 작업
- Task #11: 두 번째 하위 작업`
	result := ParsePlanOutput(output)
	if result.Type != "split" {
		t.Errorf("expected type 'split', got '%s'", result.Type)
	}
	if len(result.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(result.Children))
	}
	if result.Children[0].ID != 10 {
		t.Errorf("expected child ID 10, got %d", result.Children[0].ID)
	}
	if result.Children[1].ID != 11 {
		t.Errorf("expected child ID 11, got %d", result.Children[1].ID)
	}
}

func TestParsePlanOutput_SplitInCodeBlock(t *testing.T) {
	// This is the exact bug scenario: Claude wraps output in code blocks
	output := "```\n[SPLIT]\n- Task #37: 첫 번째\n- Task #38: 두 번째\n```"
	result := ParsePlanOutput(output)
	if result.Type != "split" {
		t.Errorf("expected type 'split', got '%s'", result.Type)
	}
	if len(result.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(result.Children))
	}
}

func TestParsePlanOutput_PlannedInCodeBlock(t *testing.T) {
	output := "```\n[PLANNED]\n## 구현 방향\n간단한 수정\n```"
	result := ParsePlanOutput(output)
	if result.Type != "planned" {
		t.Errorf("expected type 'planned', got '%s'", result.Type)
	}
	if result.Plan == "" {
		t.Error("expected non-empty plan")
	}
}

func TestParsePlanOutput_SplitInCodeBlockWithLanguage(t *testing.T) {
	// Code block with language specifier: ```text
	output := "```text\n[SPLIT]\n- Task #5: 작업\n```"
	result := ParsePlanOutput(output)
	if result.Type != "split" {
		t.Errorf("expected type 'split', got '%s'", result.Type)
	}
	if len(result.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(result.Children))
	}
}

func TestParsePlanOutput_SplitWithCRLF(t *testing.T) {
	// \r\n line endings (Windows-style from Claude)
	output := "\r\n[SPLIT]\r\n- Task #20: 작업 A\r\n- Task #21: 작업 B\r\n"
	result := ParsePlanOutput(output)
	if result.Type != "split" {
		t.Errorf("expected type 'split', got '%s'", result.Type)
	}
	if len(result.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(result.Children))
	}
}

func TestParsePlanOutput_SplitWithPreamble(t *testing.T) {
	// Claude adds explanation text before the marker
	output := "분석 결과 이 작업은 분할이 필요합니다.\n\n[SPLIT]\n- Task #15: 하위 작업"
	result := ParsePlanOutput(output)
	if result.Type != "split" {
		t.Errorf("expected type 'split', got '%s'", result.Type)
	}
	if len(result.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(result.Children))
	}
}

func TestParsePlanOutput_DefaultFallback(t *testing.T) {
	// No markers at all — backward compatibility
	output := "이건 그냥 자유 형식 출력입니다."
	result := ParsePlanOutput(output)
	if result.Type != "planned" {
		t.Errorf("expected type 'planned', got '%s'", result.Type)
	}
	if result.Plan != output {
		t.Errorf("expected plan to equal output")
	}
}

func TestStripCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no code block",
			input:    "[SPLIT]\n- Task #1: test",
			expected: "[SPLIT]\n- Task #1: test",
		},
		{
			name:     "simple code block",
			input:    "```\n[SPLIT]\n- Task #1: test\n```",
			expected: "[SPLIT]\n- Task #1: test",
		},
		{
			name:     "code block with language",
			input:    "```text\n[PLANNED]\nplan content\n```",
			expected: "[PLANNED]\nplan content",
		},
		{
			name:     "code block with trailing whitespace",
			input:    "```\n[SPLIT]\n- Task #1: test\n```  \n",
			expected: "[SPLIT]\n- Task #1: test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripCodeBlocks(tt.input)
			if got != tt.expected {
				t.Errorf("stripCodeBlocks(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractMarker(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMarker string
		wantFound  bool
	}{
		{
			name:       "split at start",
			input:      "[SPLIT]\n- Task #1: test",
			wantMarker: "[SPLIT]",
			wantFound:  true,
		},
		{
			name:       "planned at start",
			input:      "[PLANNED]\n## plan",
			wantMarker: "[PLANNED]",
			wantFound:  true,
		},
		{
			name:       "split with preamble",
			input:      "Some text before\n[SPLIT]\n- Task #1: test",
			wantMarker: "[SPLIT]",
			wantFound:  true,
		},
		{
			name:       "no marker",
			input:      "just some text",
			wantMarker: "",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marker, _, found := extractMarker(tt.input)
			if found != tt.wantFound {
				t.Errorf("extractMarker(%q) found = %v, want %v", tt.input, found, tt.wantFound)
			}
			if marker != tt.wantMarker {
				t.Errorf("extractMarker(%q) marker = %q, want %q", tt.input, marker, tt.wantMarker)
			}
		})
	}
}
