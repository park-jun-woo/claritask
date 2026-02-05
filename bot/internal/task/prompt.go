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
	TaskID       int
	Title        string
	Spec         string
	ParentID     *int
	Depth        int
	MaxDepth     int
	RelatedTasks []Task
	ReportPath   string
}

// BuildPlanPrompt builds prompt for Plan generation (1회차 순회)
func BuildPlanPrompt(t *Task, relatedTasks []Task, reportPath string) string {
	// Load template from prompts
	tmplContent, err := prompts.Get(prompts.DevPlatform, "task")
	if err != nil {
		// Fallback to simple prompt if template not found
		return buildSimplePlanPrompt(t, relatedTasks)
	}

	tmpl, err := template.New("plan").Parse(tmplContent)
	if err != nil {
		return buildSimplePlanPrompt(t, relatedTasks)
	}

	data := PlanPromptData{
		TaskID:       t.ID,
		Title:        t.Title,
		Spec:         t.Spec,
		ParentID:     t.ParentID,
		Depth:        t.Depth,
		MaxDepth:     MaxDepth,
		RelatedTasks: relatedTasks,
		ReportPath:   reportPath,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return buildSimplePlanPrompt(t, relatedTasks)
	}

	return buf.String()
}

// buildSimplePlanPrompt is fallback when template fails
func buildSimplePlanPrompt(t *Task, relatedTasks []Task) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Task: %s\n\n", t.Title))
	sb.WriteString("## 요구사항\n")
	sb.WriteString(t.Spec)
	sb.WriteString("\n\n")

	if len(relatedTasks) > 0 {
		sb.WriteString("## 연관 자료\n\n")
		for _, rt := range relatedTasks {
			sb.WriteString(fmt.Sprintf("### Task #%d: %s\n", rt.ID, rt.Title))
			sb.WriteString(fmt.Sprintf("**명세서**: %s\n\n", rt.Spec))
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString("위 요구사항과 연관 자료를 참고하여 [SUBDIVIDED] 또는 [PLANNED] 형식으로 응답하세요.\n")

	return sb.String()
}

// BuildExecutePrompt builds prompt for execution (2회차 순회)
func BuildExecutePrompt(t *Task, relatedTasks []Task, reportPath string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Task: %s\n\n", t.Title))
	sb.WriteString("## 계획서\n")
	sb.WriteString(t.Plan)
	sb.WriteString("\n\n")

	if len(relatedTasks) > 0 {
		sb.WriteString("## 연관 자료\n\n")
		for _, rt := range relatedTasks {
			sb.WriteString(fmt.Sprintf("### Task #%d: %s\n", rt.ID, rt.Title))
			sb.WriteString(fmt.Sprintf("**계획서**: %s\n\n", rt.Plan))
		}
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
