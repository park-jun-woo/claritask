package task

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

const taskDirName = "tasks"

// TaskDir returns the task directory path: {projectPath}/.claribot/tasks/
func TaskDir(projectPath string) string {
	return filepath.Join(projectPath, ".claribot", taskDirName)
}

// TaskFilePath returns the task file path: {projectPath}/.claribot/tasks/{id}.md
func TaskFilePath(projectPath string, id int) string {
	return filepath.Join(TaskDir(projectPath), fmt.Sprintf("%d.md", id))
}

// PlanFilePath returns the plan file path: {projectPath}/.claribot/tasks/{id}.plan.md
func PlanFilePath(projectPath string, id int) string {
	return filepath.Join(TaskDir(projectPath), fmt.Sprintf("%d.plan.md", id))
}

// ReportFilePath returns the report file path: {projectPath}/.claribot/tasks/{id}.report.md
func ReportFilePath(projectPath string, id int) string {
	return filepath.Join(TaskDir(projectPath), fmt.Sprintf("%d.report.md", id))
}

// ErrorFilePath returns the error file path: {projectPath}/.claribot/tasks/{id}.error.md
func ErrorFilePath(projectPath string, id int) string {
	return filepath.Join(TaskDir(projectPath), fmt.Sprintf("%d.error.md", id))
}

// EnsureTaskDir creates the task directory if it doesn't exist.
func EnsureTaskDir(projectPath string) error {
	dir := TaskDir(projectPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create task dir: %w", err)
	}
	return nil
}

// TaskContent represents parsed content of a task .md file.
type TaskContent struct {
	Frontmatter Frontmatter
	Title       string
	Body        string
}

// ReadTaskContent reads and parses a task .md file.
func ReadTaskContent(projectPath string, id int) (*TaskContent, error) {
	path := TaskFilePath(projectPath, id)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read task file: %w", err)
	}

	fm, title, body, err := ParseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse task file %s: %w", path, err)
	}

	return &TaskContent{
		Frontmatter: fm,
		Title:       title,
		Body:        body,
	}, nil
}

// WriteTaskContent writes a task .md file with frontmatter.
func WriteTaskContent(projectPath string, id int, fm Frontmatter, title, body string) error {
	if err := EnsureTaskDir(projectPath); err != nil {
		return err
	}

	content := FormatFrontmatter(fm, title, body)
	path := TaskFilePath(projectPath, id)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write task file: %w", err)
	}
	return nil
}

// ReadPlanContent reads a plan .md file. Returns ("", nil) if file doesn't exist.
func ReadPlanContent(projectPath string, id int) (string, error) {
	path := PlanFilePath(projectPath, id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read plan file: %w", err)
	}
	return string(data), nil
}

// WritePlanContent writes a plan .md file.
func WritePlanContent(projectPath string, id int, content string) error {
	if err := EnsureTaskDir(projectPath); err != nil {
		return err
	}

	path := PlanFilePath(projectPath, id)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write plan file: %w", err)
	}
	return nil
}

// ReadReportContent reads a report .md file. Returns ("", nil) if file doesn't exist.
func ReadReportContent(projectPath string, id int) (string, error) {
	path := ReportFilePath(projectPath, id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read report file: %w", err)
	}
	return string(data), nil
}

// updateTaskFileStatus reads existing task.md, updates status, and writes back.
// If file doesn't exist, does nothing (returns nil).
func updateTaskFileStatus(projectPath string, id int, status string) error {
	tc, err := ReadTaskContent(projectPath, id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		// Check unwrapped too (ReadTaskContent wraps the error)
		path := TaskFilePath(projectPath, id)
		if _, statErr := os.Stat(path); errors.Is(statErr, os.ErrNotExist) {
			return nil
		}
		return err
	}
	tc.Frontmatter.Status = status
	return WriteTaskContent(projectPath, id, tc.Frontmatter, tc.Title, tc.Body)
}

// CheckAndRestoreTaskFile checks if a task file exists; if not, attempts git restore.
func CheckAndRestoreTaskFile(projectPath string, taskID int) {
	path := TaskFilePath(projectPath, taskID)
	if _, err := os.Stat(path); err == nil {
		return // 파일 존재
	}
	relPath, err := filepath.Rel(projectPath, path)
	if err != nil {
		return
	}
	if err := GitRestore(projectPath, relPath); err == nil {
		log.Printf("[Task] task 파일 복구됨 (#%d) via git restore", taskID)
	}
}

// WriteReportContent writes a report .md file.
func WriteReportContent(projectPath string, id int, content string) error {
	if err := EnsureTaskDir(projectPath); err != nil {
		return err
	}

	path := ReportFilePath(projectPath, id)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write report file: %w", err)
	}
	return nil
}

// ReadErrorContent reads an error .md file. Returns ("", nil) if file doesn't exist.
func ReadErrorContent(projectPath string, id int) (string, error) {
	path := ErrorFilePath(projectPath, id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read error file: %w", err)
	}
	return string(data), nil
}

// WriteErrorContent writes an error .md file.
func WriteErrorContent(projectPath string, id int, content string) error {
	if err := EnsureTaskDir(projectPath); err != nil {
		return err
	}

	path := ErrorFilePath(projectPath, id)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write error file: %w", err)
	}
	return nil
}

var taskFileRe = regexp.MustCompile(`^(\d+)\.md$`)

// ScanTaskFiles scans the task directory and returns a map of task ID → file path.
func ScanTaskFiles(projectPath string) (map[int]string, error) {
	dir := TaskDir(projectPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[int]string{}, nil
		}
		return nil, fmt.Errorf("read task dir: %w", err)
	}

	result := make(map[int]string)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := taskFileRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		id, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		result[id] = filepath.Join(dir, e.Name())
	}
	return result, nil
}

// computeDepth calculates the depth of a task by walking the parent chain.
// parentMap maps task ID → parent_id (nil means root).
func computeDepth(id int, parentMap map[int]*int) int {
	visited := make(map[int]bool)
	depth := 0
	cur := id
	for {
		pid, ok := parentMap[cur]
		if !ok || pid == nil {
			return depth
		}
		depth++
		if depth > MaxDepth+1 {
			return depth // circular guard
		}
		if visited[*pid] {
			return depth // circular ref
		}
		visited[cur] = true
		cur = *pid
	}
}

// LoadContent populates a Task's content fields (Spec, Plan, Report, Error) from files.
// Fields are only overwritten if the file exists and has non-empty content.
func LoadContent(projectPath string, t *Task) {
	if tc, err := ReadTaskContent(projectPath, t.ID); err == nil && tc.Body != "" {
		t.Spec = tc.Body
	}
	if plan, err := ReadPlanContent(projectPath, t.ID); err == nil && plan != "" {
		t.Plan = plan
	}
	if report, err := ReadReportContent(projectPath, t.ID); err == nil && report != "" {
		t.Report = report
	}
	if errContent, err := ReadErrorContent(projectPath, t.ID); err == nil && errContent != "" {
		t.Error = errContent
	}
}
