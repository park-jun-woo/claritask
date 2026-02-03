package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// LLMRequest represents a request to the LLM
type LLMRequest struct {
	Prompt  string
	Timeout time.Duration // Default 60 seconds
	Retries int           // Default 3 retries
}

// LLMResponse represents a response from the LLM
type LLMResponse struct {
	Output   string
	Success  bool
	Error    string
	Duration time.Duration
}

// ContextAnalysisResult represents the result of context analysis
type ContextAnalysisResult struct {
	Tech    map[string]interface{} `json:"tech"`
	Design  map[string]interface{} `json:"design"`
	Context map[string]interface{} `json:"context"`
}

// Prompt templates
const ContextAnalysisPromptTemplate = `당신은 프로젝트 분석 전문가입니다. 아래 프로젝트 정보를 분석하여 기술 스택, 설계 패턴, 프로젝트 컨텍스트를 JSON 형식으로 출력하세요.

## 프로젝트 설명
%s

%s

## 출력 형식
다음 JSON 형식으로 응답하세요:
` + "```json" + `
{
  "tech": {
    "language": "프로그래밍 언어",
    "framework": "사용 프레임워크",
    "database": "데이터베이스 (있다면)",
    "build_tool": "빌드 도구",
    "dependencies": ["주요 의존성 목록"]
  },
  "design": {
    "architecture": "아키텍처 패턴 (MVC, 클린 아키텍처, 레이어드 등)",
    "patterns": ["사용된 디자인 패턴"],
    "modules": ["주요 모듈/패키지"]
  },
  "context": {
    "project_type": "프로젝트 유형 (CLI, API, 웹앱 등)",
    "domain": "도메인 영역",
    "target_users": "대상 사용자"
  }
}
` + "```"

const SpecsGenerationPromptTemplate = `당신은 소프트웨어 스펙 문서 작성 전문가입니다. 아래 프로젝트 정보를 바탕으로 상세한 스펙 문서를 Markdown 형식으로 작성하세요.

## 프로젝트 정보
- **ID**: %s
- **이름**: %s
- **설명**: %s

## 기술 스택
%s

## 설계
%s

## 컨텍스트
%s

## 작성 요구사항
1. 프로젝트 개요
2. 기능 명세 (Feature 목록)
3. 기술 스택 상세
4. 아키텍처 설계
5. 데이터 모델 (있다면)
6. API 명세 (있다면)
7. 개발 가이드라인

Markdown 문서로 출력하세요.`

const SpecsRevisionPromptTemplate = `아래 스펙 문서에 대한 피드백을 반영하여 수정된 버전을 작성하세요.

## 현재 스펙 문서
%s

## 피드백
%s

## 요구사항
- 피드백을 반영하여 스펙 문서를 수정하세요
- 전체 문서를 Markdown 형식으로 출력하세요`

// CallClaude calls the claude CLI with --print flag
func CallClaude(request LLMRequest) (*LLMResponse, error) {
	if request.Timeout == 0 {
		request.Timeout = 60 * time.Second
	}
	if request.Retries == 0 {
		request.Retries = 3
	}

	var lastErr error
	startTime := time.Now()

	for attempt := 0; attempt < request.Retries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), request.Timeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "claude", "--print", request.Prompt)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		duration := time.Since(startTime)

		if err == nil {
			return &LLMResponse{
				Output:   stdout.String(),
				Success:  true,
				Duration: duration,
			}, nil
		}

		lastErr = err
		if ctx.Err() == context.DeadlineExceeded {
			lastErr = fmt.Errorf("timeout after %v", request.Timeout)
		}

		// Wait before retry
		if attempt < request.Retries-1 {
			time.Sleep(time.Second * time.Duration(attempt+1))
		}
	}

	return &LLMResponse{
		Output:   "",
		Success:  false,
		Error:    lastErr.Error(),
		Duration: time.Since(startTime),
	}, lastErr
}

// ParseContextAnalysis extracts and parses JSON from LLM output
func ParseContextAnalysis(output string) (*ContextAnalysisResult, error) {
	// Try to find JSON code block
	jsonRegex := regexp.MustCompile("(?s)```json\\s*\\n(.+?)\\n```")
	matches := jsonRegex.FindStringSubmatch(output)

	var jsonStr string
	if len(matches) >= 2 {
		jsonStr = matches[1]
	} else {
		// Try to find JSON without code block
		jsonStr = extractJSON(output)
		if jsonStr == "" {
			return nil, fmt.Errorf("no JSON found in output")
		}
	}

	var result ContextAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// extractJSON tries to extract JSON from raw text
func extractJSON(text string) string {
	// Find first { and last }
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || start >= end {
		return ""
	}
	return text[start : end+1]
}

// ParseSpecsDocument extracts markdown document from LLM output
func ParseSpecsDocument(output string) (string, error) {
	// Try to find markdown code block
	mdRegex := regexp.MustCompile("(?s)```markdown\\s*\\n(.+?)\\n```")
	matches := mdRegex.FindStringSubmatch(output)

	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1]), nil
	}

	// Try to find generic code block
	codeRegex := regexp.MustCompile("(?s)```\\s*\\n(.+?)\\n```")
	matches = codeRegex.FindStringSubmatch(output)

	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1]), nil
	}

	// Return entire output if no code block found
	return strings.TrimSpace(output), nil
}

// BuildContextAnalysisPrompt builds a prompt for context analysis
func BuildContextAnalysisPrompt(scanResult *ScanResult, description string) string {
	scanResultStr := FormatScanResultForLLM(scanResult)
	return fmt.Sprintf(ContextAnalysisPromptTemplate, description, scanResultStr)
}

// BuildSpecsGenerationPrompt builds a prompt for specs generation
func BuildSpecsGenerationPrompt(projectID, name, description string,
	tech, design, context map[string]interface{}) string {

	techJSON, _ := json.MarshalIndent(tech, "", "  ")
	designJSON, _ := json.MarshalIndent(design, "", "  ")
	contextJSON, _ := json.MarshalIndent(context, "", "  ")

	return fmt.Sprintf(SpecsGenerationPromptTemplate,
		projectID, name, description,
		string(techJSON), string(designJSON), string(contextJSON))
}

// BuildSpecsRevisionPrompt builds a prompt for specs revision
func BuildSpecsRevisionPrompt(currentSpecs, feedback string) string {
	return fmt.Sprintf(SpecsRevisionPromptTemplate, currentSpecs, feedback)
}
