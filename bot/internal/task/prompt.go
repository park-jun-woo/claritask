package task

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"parkjunwoo.com/claribot/internal/prompts"
)

// PlanPromptData holds data for plan prompt template
type PlanPromptData struct {
	TaskID     int
	Title      string
	Spec       string
	ParentID   *int
	Depth      int
	MaxDepth   int
	ContextMap string
	ReportPath string
}

// BuildPlanPrompt builds prompt for Plan generation (1회차 순회)
func BuildPlanPrompt(t *Task, contextMap string, reportPath string) string {
	// Load template from prompts
	tmplContent, err := prompts.Get(prompts.Common, "task")
	if err != nil {
		// Fallback to simple prompt if template not found
		return buildSimplePlanPrompt(t, contextMap)
	}

	tmpl, err := template.New("plan").Parse(tmplContent)
	if err != nil {
		return buildSimplePlanPrompt(t, contextMap)
	}

	data := PlanPromptData{
		TaskID:     t.ID,
		Title:      t.Title,
		Spec:       t.Spec,
		ParentID:   t.ParentID,
		Depth:      t.Depth,
		MaxDepth:   MaxDepth,
		ContextMap: contextMap,
		ReportPath: reportPath,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return buildSimplePlanPrompt(t, contextMap)
	}

	return buf.String()
}

// buildSimplePlanPrompt is fallback when template fails
func buildSimplePlanPrompt(t *Task, contextMap string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Task: %s\n\n", t.Title))
	sb.WriteString("## 요구사항\n")
	sb.WriteString(t.Spec)
	sb.WriteString("\n\n")

	if contextMap != "" {
		sb.WriteString("## Context Map\n\n")
		sb.WriteString("```\n")
		sb.WriteString(contextMap)
		sb.WriteString("```\n\n")
		sb.WriteString("## 조회 명령어\n\n")
		sb.WriteString("- `clari task get <id>` - 특정 Task 상세 조회 (spec, plan, report)\n")
		sb.WriteString("- `clari task list [parent_id]` - Task 목록 조회\n\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString("위 요구사항과 Context Map을 참고하여 [SPLIT] 또는 [PLANNED] 형식으로 응답하세요.\n")

	return sb.String()
}

// ExecutePromptData holds data for execute prompt template
type ExecutePromptData struct {
	TaskID     int
	Title      string
	Plan       string
	ContextMap string
	ReportPath string
}

// BuildExecutePrompt builds prompt for execution (2회차 순회)
func BuildExecutePrompt(t *Task, contextMap string, reportPath string) string {
	// Load template from prompts
	tmplContent, err := prompts.Get(prompts.Common, "task_run")
	if err != nil {
		// Fallback to simple prompt if template not found
		return buildSimpleExecutePrompt(t, contextMap, reportPath)
	}

	tmpl, err := template.New("execute").Parse(tmplContent)
	if err != nil {
		return buildSimpleExecutePrompt(t, contextMap, reportPath)
	}

	data := ExecutePromptData{
		TaskID:     t.ID,
		Title:      t.Title,
		Plan:       t.Plan,
		ContextMap: contextMap,
		ReportPath: reportPath,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return buildSimpleExecutePrompt(t, contextMap, reportPath)
	}

	return buf.String()
}

// buildSimpleExecutePrompt is fallback when template fails
func buildSimpleExecutePrompt(t *Task, contextMap string, reportPath string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Task: %s\n\n", t.Title))
	sb.WriteString("## 계획서\n")
	sb.WriteString(t.Plan)
	sb.WriteString("\n\n")

	if contextMap != "" {
		sb.WriteString("## Context Map\n\n")
		sb.WriteString("```\n")
		sb.WriteString(contextMap)
		sb.WriteString("```\n\n")
		sb.WriteString("## 조회 명령어\n\n")
		sb.WriteString("- `clari task get <id>` - 특정 Task 상세 조회 (spec, plan, report)\n")
		sb.WriteString("- `clari task list [parent_id]` - Task 목록 조회\n\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString("위 계획서와 연관 자료를 참고하여 작업을 수행하세요.\n\n")
	sb.WriteString("완료 후 보고서를 작성하세요:\n")
	sb.WriteString("- 수행한 작업 요약\n")
	sb.WriteString("- 변경된 파일 목록\n")
	sb.WriteString("- 특이사항\n\n")
	sb.WriteString("## ⚠️ 결과 보고서 파일 저장 (필수)\n\n")
	sb.WriteString(fmt.Sprintf("**모든 작업이 완료되면 반드시** 보고서를 다음 경로에 파일로 저장하세요:\n\n"))
	sb.WriteString(fmt.Sprintf("```\n파일 경로: %s\n```\n\n", reportPath))
	sb.WriteString("- 이 파일이 생성되어야 작업 완료로 인식됩니다\n")
	sb.WriteString("- 파일이 없으면 작업이 완료되지 않은 것으로 간주합니다\n")

	return sb.String()
}
