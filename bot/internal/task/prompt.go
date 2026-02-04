package task

import (
	"fmt"
	"strings"
)

// BuildPlanPrompt builds prompt for Plan generation (1회차 순회)
func BuildPlanPrompt(t *Task, relatedTasks []Task) string {
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
	sb.WriteString("위 요구사항과 연관 자료를 참고하여 실행 계획서를 작성하세요.\n\n")
	sb.WriteString("계획서에는 다음 내용을 포함하세요:\n")
	sb.WriteString("- 구현 방향\n")
	sb.WriteString("- 주요 변경 파일\n")
	sb.WriteString("- 의존성 또는 주의사항\n")

	return sb.String()
}

// BuildExecutePrompt builds prompt for execution (2회차 순회)
func BuildExecutePrompt(t *Task, relatedTasks []Task) string {
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
	sb.WriteString("- 특이사항\n")

	return sb.String()
}
