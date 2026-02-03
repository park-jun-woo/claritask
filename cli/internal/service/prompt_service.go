package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// PromptOption represents a menu option for user prompts
type PromptOption struct {
	Key         string
	Label       string
	Description string
}

// ApprovalResult represents the result of an approval prompt
type ApprovalResult struct {
	Action    string // "approve", "edit", "reanalyze", "cancel"
	EditKey   string // Key to edit (e.g., "tech.database")
	EditValue string // New value
}

// DefaultReader is the default input reader (can be replaced for testing)
var DefaultReader io.Reader = os.Stdin

// DefaultWriter is the default output writer (can be replaced for testing)
var DefaultWriter io.Writer = os.Stdout

// PrintBox prints content in a box format
func PrintBox(title string, content string) {
	width := 60
	border := strings.Repeat("═", width-2)

	fmt.Fprintf(DefaultWriter, "╔%s╗\n", border)
	fmt.Fprintf(DefaultWriter, "║ %-*s ║\n", width-4, title)
	fmt.Fprintf(DefaultWriter, "╠%s╣\n", border)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Truncate long lines
		if len(line) > width-4 {
			line = line[:width-7] + "..."
		}
		fmt.Fprintf(DefaultWriter, "║ %-*s ║\n", width-4, line)
	}

	fmt.Fprintf(DefaultWriter, "╚%s╝\n", border)
}

// PrintAnalysisResult prints the analysis result in a formatted way
func PrintAnalysisResult(tech, design, context map[string]interface{}) {
	fmt.Fprintf(DefaultWriter, "\n")
	PrintBox("분석 결과", "LLM이 프로젝트를 분석했습니다.")
	fmt.Fprintf(DefaultWriter, "\n")

	// Tech Stack
	fmt.Fprintf(DefaultWriter, "📦 기술 스택 (Tech Stack)\n")
	fmt.Fprintf(DefaultWriter, "%s\n", strings.Repeat("─", 40))
	printMap(tech, "  ")
	fmt.Fprintf(DefaultWriter, "\n")

	// Design
	fmt.Fprintf(DefaultWriter, "🏗️  설계 (Design)\n")
	fmt.Fprintf(DefaultWriter, "%s\n", strings.Repeat("─", 40))
	printMap(design, "  ")
	fmt.Fprintf(DefaultWriter, "\n")

	// Context
	fmt.Fprintf(DefaultWriter, "📋 컨텍스트 (Context)\n")
	fmt.Fprintf(DefaultWriter, "%s\n", strings.Repeat("─", 40))
	printMap(context, "  ")
	fmt.Fprintf(DefaultWriter, "\n")
}

// printMap recursively prints a map with indentation
func printMap(m map[string]interface{}, indent string) {
	for key, value := range m {
		switch v := value.(type) {
		case map[string]interface{}:
			fmt.Fprintf(DefaultWriter, "%s%s:\n", indent, key)
			printMap(v, indent+"  ")
		case []interface{}:
			fmt.Fprintf(DefaultWriter, "%s%s:\n", indent, key)
			for _, item := range v {
				fmt.Fprintf(DefaultWriter, "%s  - %v\n", indent, item)
			}
		default:
			fmt.Fprintf(DefaultWriter, "%s%s: %v\n", indent, key, value)
		}
	}
}

// PromptApproval displays options and waits for user selection
func PromptApproval(options []PromptOption) (*ApprovalResult, error) {
	fmt.Fprintf(DefaultWriter, "\n선택하세요:\n")
	for _, opt := range options {
		fmt.Fprintf(DefaultWriter, "  [%s] %s", opt.Key, opt.Label)
		if opt.Description != "" {
			fmt.Fprintf(DefaultWriter, " - %s", opt.Description)
		}
		fmt.Fprintf(DefaultWriter, "\n")
	}
	fmt.Fprintf(DefaultWriter, "\n> ")

	scanner := bufio.NewScanner(DefaultReader)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("no input")
	}

	input := strings.ToLower(strings.TrimSpace(scanner.Text()))

	// Map input to action
	for _, opt := range options {
		if strings.ToLower(opt.Key) == input {
			return &ApprovalResult{Action: strings.ToLower(opt.Label)}, nil
		}
	}

	// Check common actions
	switch input {
	case "a", "approve", "승인":
		return &ApprovalResult{Action: "approve"}, nil
	case "e", "edit", "수정":
		return &ApprovalResult{Action: "edit"}, nil
	case "r", "reanalyze", "재분석":
		return &ApprovalResult{Action: "reanalyze"}, nil
	case "q", "quit", "cancel", "취소":
		return &ApprovalResult{Action: "cancel"}, nil
	}

	return nil, fmt.Errorf("invalid option: %s", input)
}

// GetStandardApprovalOptions returns standard approval options
func GetStandardApprovalOptions() []PromptOption {
	return []PromptOption{
		{Key: "A", Label: "approve", Description: "승인하고 다음 단계로"},
		{Key: "E", Label: "edit", Description: "항목 수정"},
		{Key: "R", Label: "reanalyze", Description: "재분석 요청"},
		{Key: "Q", Label: "cancel", Description: "취소"},
	}
}

// PromptEdit prompts user to edit a specific key in the data
func PromptEdit(currentData map[string]interface{}) (string, string, error) {
	// Show current data
	fmt.Fprintf(DefaultWriter, "\n현재 데이터:\n")
	dataJSON, _ := json.MarshalIndent(currentData, "", "  ")
	fmt.Fprintf(DefaultWriter, "%s\n", string(dataJSON))

	// Ask for key to edit
	fmt.Fprintf(DefaultWriter, "\n수정할 항목 키 (예: language, framework): ")
	scanner := bufio.NewScanner(DefaultReader)

	if !scanner.Scan() {
		return "", "", fmt.Errorf("no input")
	}
	key := strings.TrimSpace(scanner.Text())

	// Show current value if exists
	if val, ok := currentData[key]; ok {
		fmt.Fprintf(DefaultWriter, "현재 값: %v\n", val)
	}

	// Ask for new value
	fmt.Fprintf(DefaultWriter, "새 값: ")
	if !scanner.Scan() {
		return "", "", fmt.Errorf("no input")
	}
	value := strings.TrimSpace(scanner.Text())

	return key, value, nil
}

// PromptMultilineInput prompts for multi-line input (ends with two empty lines)
func PromptMultilineInput(prompt string) (string, error) {
	fmt.Fprintf(DefaultWriter, "%s\n", prompt)
	fmt.Fprintf(DefaultWriter, "(입력을 마치려면 빈 줄을 두 번 입력하세요)\n\n")

	scanner := bufio.NewScanner(DefaultReader)
	var lines []string
	emptyCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			emptyCount++
			if emptyCount >= 2 {
				break
			}
			lines = append(lines, line)
		} else {
			emptyCount = 0
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	// Remove trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n"), nil
}

// PromptConfirm prompts for yes/no confirmation
func PromptConfirm(message string, defaultYes bool) (bool, error) {
	suffix := "[y/N]"
	if defaultYes {
		suffix = "[Y/n]"
	}

	fmt.Fprintf(DefaultWriter, "%s %s: ", message, suffix)

	scanner := bufio.NewScanner(DefaultReader)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return defaultYes, nil
	}

	input := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if input == "" {
		return defaultYes, nil
	}

	switch input {
	case "y", "yes", "예":
		return true, nil
	case "n", "no", "아니오":
		return false, nil
	default:
		return defaultYes, nil
	}
}

// PrintSpecs prints the specs document with pagination if needed
func PrintSpecs(specs string) {
	fmt.Fprintf(DefaultWriter, "\n")
	fmt.Fprintf(DefaultWriter, "═══════════════════════════════════════════════════════════\n")
	fmt.Fprintf(DefaultWriter, "                    📄 Specs 문서\n")
	fmt.Fprintf(DefaultWriter, "═══════════════════════════════════════════════════════════\n")
	fmt.Fprintf(DefaultWriter, "\n%s\n", specs)
	fmt.Fprintf(DefaultWriter, "\n═══════════════════════════════════════════════════════════\n")
}

// PrintFinalResult prints the final completion message
func PrintFinalResult(projectID, dbPath, specsPath string) {
	content := fmt.Sprintf(`프로젝트 ID: %s
데이터베이스: %s
스펙 문서: %s

초기화가 완료되었습니다!
다음 명령으로 프로젝트를 관리하세요:

  clari project get %s
  clari feature list
  clari task list`, projectID, dbPath, specsPath, projectID)

	fmt.Fprintf(DefaultWriter, "\n")
	PrintBox("✅ 초기화 완료", content)
}

// PrintProgress prints a progress message
func PrintProgress(phase int, total int, message string) {
	fmt.Fprintf(DefaultWriter, "\n[%d/%d] %s\n", phase, total, message)
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Fprintf(DefaultWriter, "\n❌ 오류: %s\n", message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Fprintf(DefaultWriter, "\n⚠️  경고: %s\n", message)
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	fmt.Fprintf(DefaultWriter, "\nℹ️  %s\n", message)
}
