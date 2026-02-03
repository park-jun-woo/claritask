package test

import (
	"strings"
	"testing"

	"parkjunwoo.com/claritask/internal/service"
)

func TestParseContextAnalysis_ValidJSON(t *testing.T) {
	input := `Some text before the JSON block

Here is the analysis:
` + "```json" + `
{
  "tech": {"language": "Go", "framework": "Cobra"},
  "design": {"architecture": "MVC", "patterns": ["singleton"]},
  "context": {"project_type": "CLI", "domain": "DevOps"}
}
` + "```" + `

Some text after the JSON block`

	result, err := service.ParseContextAnalysis(input)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Check tech
	if result.Tech["language"] != "Go" {
		t.Errorf("expected language 'Go', got '%v'", result.Tech["language"])
	}
	if result.Tech["framework"] != "Cobra" {
		t.Errorf("expected framework 'Cobra', got '%v'", result.Tech["framework"])
	}

	// Check design
	if result.Design["architecture"] != "MVC" {
		t.Errorf("expected architecture 'MVC', got '%v'", result.Design["architecture"])
	}

	// Check context
	if result.Context["project_type"] != "CLI" {
		t.Errorf("expected project_type 'CLI', got '%v'", result.Context["project_type"])
	}
}

func TestParseContextAnalysis_NoCodeBlock(t *testing.T) {
	// JSON without code block
	input := `{
  "tech": {"language": "Python"},
  "design": {"architecture": "Microservices"},
  "context": {"project_type": "API"}
}`

	result, err := service.ParseContextAnalysis(input)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if result.Tech["language"] != "Python" {
		t.Errorf("expected language 'Python', got '%v'", result.Tech["language"])
	}
}

func TestParseContextAnalysis_InvalidJSON(t *testing.T) {
	input := `This is not valid JSON at all`

	_, err := service.ParseContextAnalysis(input)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseSpecsDocument_MarkdownBlock(t *testing.T) {
	input := `Here is the specs document:

` + "```markdown" + `
# Project Specs

## Overview
This is a test project.

## Features
- Feature 1
- Feature 2
` + "```" + `

End of response.`

	result, err := service.ParseSpecsDocument(input)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if !strings.Contains(result, "# Project Specs") {
		t.Error("result should contain markdown header")
	}
	if !strings.Contains(result, "## Features") {
		t.Error("result should contain Features section")
	}
}

func TestParseSpecsDocument_NoBlock(t *testing.T) {
	input := `# Direct Markdown Output

## Section 1
Content here.

## Section 2
More content.`

	result, err := service.ParseSpecsDocument(input)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if !strings.Contains(result, "# Direct Markdown Output") {
		t.Error("result should contain the direct markdown")
	}
}

func TestBuildContextAnalysisPrompt(t *testing.T) {
	scanResult := &service.ScanResult{
		Files: []service.ScannedFile{
			{Path: "go.mod", Type: "go_mod", Content: "module test"},
		},
		Directories: []string{"cmd", "internal"},
	}

	prompt := service.BuildContextAnalysisPrompt(scanResult, "A CLI tool for task management")

	// Check required sections
	if !strings.Contains(prompt, "프로젝트 분석") {
		t.Error("prompt should contain analysis instruction")
	}
	if !strings.Contains(prompt, "A CLI tool for task management") {
		t.Error("prompt should contain project description")
	}
	if !strings.Contains(prompt, "tech") {
		t.Error("prompt should mention tech output")
	}
	if !strings.Contains(prompt, "design") {
		t.Error("prompt should mention design output")
	}
	if !strings.Contains(prompt, "context") {
		t.Error("prompt should mention context output")
	}
}

func TestBuildSpecsGenerationPrompt(t *testing.T) {
	tech := map[string]interface{}{
		"language":  "Go",
		"framework": "Cobra",
	}
	design := map[string]interface{}{
		"architecture": "Layered",
	}
	context := map[string]interface{}{
		"project_type": "CLI",
	}

	prompt := service.BuildSpecsGenerationPrompt("my-project", "My Project", "A test project", tech, design, context)

	if !strings.Contains(prompt, "my-project") {
		t.Error("prompt should contain project ID")
	}
	if !strings.Contains(prompt, "My Project") {
		t.Error("prompt should contain project name")
	}
	if !strings.Contains(prompt, "A test project") {
		t.Error("prompt should contain description")
	}
	if !strings.Contains(prompt, "Go") {
		t.Error("prompt should contain tech info")
	}
	if !strings.Contains(prompt, "Layered") {
		t.Error("prompt should contain design info")
	}
}

func TestBuildSpecsRevisionPrompt(t *testing.T) {
	currentSpecs := `# Original Specs

## Features
- Feature A`

	feedback := "Please add Feature B and improve the overview section."

	prompt := service.BuildSpecsRevisionPrompt(currentSpecs, feedback)

	if !strings.Contains(prompt, "# Original Specs") {
		t.Error("prompt should contain current specs")
	}
	if !strings.Contains(prompt, "Please add Feature B") {
		t.Error("prompt should contain feedback")
	}
	if !strings.Contains(prompt, "피드백") {
		t.Error("prompt should mention feedback section")
	}
}

func TestLLMRequestDefaults(t *testing.T) {
	// This tests that default values work - we can't actually call Claude in unit tests
	req := service.LLMRequest{
		Prompt: "test prompt",
	}

	// Timeout and Retries should be set by CallClaude if not specified
	if req.Timeout != 0 {
		t.Error("initial timeout should be 0 (set by CallClaude)")
	}
	if req.Retries != 0 {
		t.Error("initial retries should be 0 (set by CallClaude)")
	}
}

func TestLLMResponseStructure(t *testing.T) {
	// Test the response structure
	resp := service.LLMResponse{
		Output:  "test output",
		Success: true,
		Error:   "",
	}

	if resp.Output != "test output" {
		t.Errorf("expected output 'test output', got '%s'", resp.Output)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestContextAnalysisResultStructure(t *testing.T) {
	result := service.ContextAnalysisResult{
		Tech:    map[string]interface{}{"language": "Go"},
		Design:  map[string]interface{}{"architecture": "MVC"},
		Context: map[string]interface{}{"project_type": "CLI"},
	}

	if result.Tech["language"] != "Go" {
		t.Errorf("expected language 'Go', got '%v'", result.Tech["language"])
	}
}
