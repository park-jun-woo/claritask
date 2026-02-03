package service

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// ExecutionState represents the current execution state
type ExecutionState struct {
	Running        bool       `json:"running"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CurrentTask    *int64     `json:"current_task,omitempty"`
	TotalTasks     int        `json:"total_tasks"`
	CompletedTasks int        `json:"completed_tasks"`
	FailedTasks    int        `json:"failed_tasks"`
}

// GetExecutionState retrieves the current execution state
func GetExecutionState(database *db.DB) (*ExecutionState, error) {
	state := &ExecutionState{}

	// Check if running
	running, _ := GetState(database, "execution_running")
	state.Running = running == "true"

	// Get started time
	startedAt, _ := GetState(database, "execution_started_at")
	if startedAt != "" {
		if t, err := time.Parse(time.RFC3339, startedAt); err == nil {
			state.StartedAt = &t
		}
	}

	// Get current task
	currentTaskStr, _ := GetState(database, "execution_current_task")
	if currentTaskStr != "" {
		if id, err := strconv.ParseInt(currentTaskStr, 10, 64); err == nil {
			state.CurrentTask = &id
		}
	}

	// Get task counts
	taskStatus, _ := GetTaskStatus(database)
	if taskStatus != nil {
		state.TotalTasks = taskStatus.Total
		state.CompletedTasks = taskStatus.Done
		state.FailedTasks = taskStatus.Failed
	}

	return state, nil
}

// StartExecution marks execution as started
func StartExecution(database *db.DB) error {
	SetState(database, "execution_running", "true")
	SetState(database, "execution_started_at", time.Now().Format(time.RFC3339))
	SetState(database, "execution_stop_requested", "false")
	return nil
}

// StopExecution marks execution as stopped
func StopExecution(database *db.DB) error {
	SetState(database, "execution_running", "false")
	SetState(database, "execution_current_task", "")
	return nil
}

// RequestStop requests execution to stop
func RequestStop(database *db.DB) error {
	SetState(database, "execution_stop_requested", "true")
	return nil
}

// IsStopRequested checks if stop has been requested
func IsStopRequested(database *db.DB) bool {
	requested, _ := GetState(database, "execution_stop_requested")
	return requested == "true"
}

// UpdateExecutionCurrentTask updates the current task being executed
func UpdateExecutionCurrentTask(database *db.DB, taskID int64) error {
	SetState(database, "execution_current_task", strconv.FormatInt(taskID, 10))
	return nil
}

// ExecutionPlan represents a planned execution
type ExecutionPlan struct {
	Tasks []PlannedTask `json:"tasks"`
	Total int           `json:"total"`
}

// PlannedTask represents a task in the execution plan
type PlannedTask struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Feature   string   `json:"feature"`
	DependsOn []string `json:"depends_on,omitempty"`
	Order     int      `json:"order"`
}

// GenerateExecutionPlan generates an execution plan (dry-run)
func GenerateExecutionPlan(database *db.DB, featureID *int64) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		Tasks: []PlannedTask{},
	}

	// Get project
	project, err := GetProject(database)
	if err != nil {
		return nil, err
	}

	// Get all features
	features, err := ListFeatures(database, project.ID)
	if err != nil {
		return nil, err
	}

	order := 0
	for _, feature := range features {
		// Filter by feature if specified
		if featureID != nil && feature.ID != *featureID {
			continue
		}

		// Get tasks for this feature
		tasks, err := ListTasksByFeature(database, feature.ID)
		if err != nil {
			continue
		}

		for _, task := range tasks {
			// Skip non-pending tasks
			if task.Status != "pending" {
				continue
			}

			order++
			planned := PlannedTask{
				ID:      strconv.FormatInt(task.ID, 10),
				Title:   task.Title,
				Feature: feature.Name,
				Order:   order,
			}

			// Get dependencies
			deps, _ := GetTaskDependencies(database, task.ID)
			for _, dep := range deps {
				planned.DependsOn = append(planned.DependsOn, strconv.FormatInt(dep.ID, 10))
			}

			plan.Tasks = append(plan.Tasks, planned)
		}
	}

	plan.Total = len(plan.Tasks)
	return plan, nil
}

// ClaudeResult represents the result of Claude execution
type ClaudeResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// ExecuteTaskWithClaude executes a task using Claude in headless mode
func ExecuteTaskWithClaude(task *model.Task, manifest *model.Manifest) (*ClaudeResult, error) {
	prompt := buildTaskPrompt(task, manifest)

	cmd := exec.Command("claude", "--print", prompt)
	output, err := cmd.CombinedOutput()

	result := &ClaudeResult{
		Output: string(output),
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		result.Success = true
	}

	return result, nil
}

// buildTaskPrompt builds a prompt for task execution
func buildTaskPrompt(task *model.Task, manifest *model.Manifest) string {
	prompt := fmt.Sprintf(`[CLARITASK EXECUTION]

Task ID: %d
Title: %s
Content:
%s

`, task.ID, task.Title, task.Content)

	if task.TargetFile != "" {
		prompt += fmt.Sprintf("Target File: %s\n", task.TargetFile)
	}
	if task.TargetFunction != "" {
		prompt += fmt.Sprintf("Target Function: %s\n", task.TargetFunction)
	}

	// Add manifest context
	if manifest != nil {
		if len(manifest.Tech) > 0 {
			techJSON, _ := json.Marshal(manifest.Tech)
			prompt += fmt.Sprintf("\nTech Stack: %s\n", string(techJSON))
		}
		if len(manifest.Design) > 0 {
			designJSON, _ := json.Marshal(manifest.Design)
			prompt += fmt.Sprintf("Design: %s\n", string(designJSON))
		}
	}

	prompt += `
---
Instructions:
1. Implement the task as described above
2. If the task involves code, write the implementation
3. When done, summarize what was accomplished
`

	return prompt
}

// ExecutionOptions represents execution options
type ExecutionOptions struct {
	FeatureID           *int64 // Execute specific feature only
	DryRun              bool   // Only show execution plan
	FallbackInteractive bool   // Switch to interactive mode on failure
}

// ExecuteAllTasks executes all pending tasks
func ExecuteAllTasks(database *db.DB, options ExecutionOptions) error {
	if err := StartExecution(database); err != nil {
		return err
	}
	defer StopExecution(database)

	for {
		// Check for stop request
		if IsStopRequested(database) {
			return nil
		}

		// Get next executable task
		response, err := PopTaskFull(database)
		if err != nil {
			return fmt.Errorf("pop task: %w", err)
		}
		if response.Task == nil {
			// All tasks completed
			return nil
		}

		task := response.Task

		// Filter by feature if specified
		if options.FeatureID != nil && task.FeatureID != *options.FeatureID {
			// Reset task and continue
			ResetTaskToPending(database, task.ID)
			continue
		}

		// Update current task
		UpdateExecutionCurrentTask(database, task.ID)

		// Execute with Claude
		result, err := ExecuteTaskWithClaude(task, &response.Manifest)
		if err != nil {
			FailTask(database, task.ID, err.Error())
			if options.FallbackInteractive {
				// TODO: Call interactive debugging
				continue
			}
			return fmt.Errorf("execute task %d: %w", task.ID, err)
		}

		// Process result
		if result.Success {
			CompleteTask(database, task.ID, result.Output)
		} else {
			if options.FallbackInteractive {
				// Mark as failed, will need interactive debugging
				FailTask(database, task.ID, result.Error)
				// Continue to next task instead of blocking
			} else {
				FailTask(database, task.ID, result.Error)
				return fmt.Errorf("task %d failed: %s", task.ID, result.Error)
			}
		}
	}
}

// ExecuteFeature executes all tasks for a specific feature
func ExecuteFeature(database *db.DB, featureID int64) error {
	// Check feature dependencies
	deps, err := GetFeatureDependencies(database, featureID)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		if dep.Status != "done" {
			return fmt.Errorf("feature %d blocked by feature %d (%s)", featureID, dep.ID, dep.Name)
		}
	}

	// Execute tasks for this feature
	options := ExecutionOptions{FeatureID: &featureID}
	return ExecuteAllTasks(database, options)
}

// ExecutionProgress represents execution progress
type ExecutionProgress struct {
	TotalFeatures     int        `json:"total_features"`
	CompletedFeatures int        `json:"completed_features"`
	TotalTasks        int        `json:"total_tasks"`
	Pending           int        `json:"pending"`
	Doing             int        `json:"doing"`
	Done              int        `json:"done"`
	Failed            int        `json:"failed"`
	Progress          float64    `json:"progress"`
	CurrentTask       *TaskInfo  `json:"current_task,omitempty"`
	FailedTasks       []TaskInfo `json:"failed_tasks,omitempty"`
}

// TaskInfo represents basic task information
type TaskInfo struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// GetExecutionProgress retrieves execution progress
func GetExecutionProgress(database *db.DB) (*ExecutionProgress, error) {
	progress := &ExecutionProgress{}

	// Get project
	project, err := GetProject(database)
	if err != nil {
		return nil, err
	}

	// Get feature counts
	features, err := ListFeatures(database, project.ID)
	if err == nil {
		progress.TotalFeatures = len(features)
		for _, f := range features {
			if f.Status == "done" {
				progress.CompletedFeatures++
			}
		}
	}

	// Get task counts
	taskStatus, err := GetTaskStatus(database)
	if err != nil {
		return nil, err
	}

	progress.TotalTasks = taskStatus.Total
	progress.Pending = taskStatus.Pending
	progress.Doing = taskStatus.Doing
	progress.Done = taskStatus.Done
	progress.Failed = taskStatus.Failed
	progress.Progress = taskStatus.Progress

	// Get current task
	currentTaskStr, _ := GetState(database, "execution_current_task")
	if currentTaskStr != "" {
		if id, err := strconv.ParseInt(currentTaskStr, 10, 64); err == nil {
			task, err := GetTask(database, id)
			if err == nil {
				progress.CurrentTask = &TaskInfo{
					ID:     strconv.FormatInt(task.ID, 10),
					Title:  task.Title,
					Status: task.Status,
				}
			}
		}
	}

	// Get failed tasks
	rows, err := database.Query(
		`SELECT id, title, error FROM tasks WHERE status = 'failed' ORDER BY id`,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			var title, errMsg string
			if err := rows.Scan(&id, &title, &errMsg); err == nil {
				progress.FailedTasks = append(progress.FailedTasks, TaskInfo{
					ID:     strconv.FormatInt(id, 10),
					Title:  title,
					Status: "failed",
					Error:  errMsg,
				})
			}
		}
	}

	return progress, nil
}
