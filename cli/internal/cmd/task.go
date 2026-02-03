package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"parkjunwoo.com/claritask/internal/service"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task management commands",
}

var taskPushCmd = &cobra.Command{
	Use:   "push '<json>'",
	Short: "Add a new task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskPush,
}

var taskPopCmd = &cobra.Command{
	Use:   "pop",
	Short: "Get next pending task with manifest",
	RunE:  runTaskPop,
}

var taskStartCmd = &cobra.Command{
	Use:   "start <task_id>",
	Short: "Start a task (pending -> doing)",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskStart,
}

var taskCompleteCmd = &cobra.Command{
	Use:   "complete <task_id> '<json>'",
	Short: "Complete a task (doing -> done)",
	Args:  cobra.ExactArgs(2),
	RunE:  runTaskComplete,
}

var taskFailCmd = &cobra.Command{
	Use:   "fail <task_id> '<json>'",
	Short: "Fail a task (doing -> failed)",
	Args:  cobra.ExactArgs(2),
	RunE:  runTaskFail,
}

var taskStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show task progress",
	RunE:  runTaskStatus,
}

var taskGetCmd = &cobra.Command{
	Use:   "get <task_id>",
	Short: "Get task by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskGet,
}

var taskListCmd = &cobra.Command{
	Use:   "list [feature_id]",
	Short: "List tasks (optionally by feature)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runTaskList,
}

func init() {
	taskCmd.AddCommand(taskPushCmd)
	taskCmd.AddCommand(taskPopCmd)
	taskCmd.AddCommand(taskStartCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	taskCmd.AddCommand(taskFailCmd)
	taskCmd.AddCommand(taskStatusCmd)
	taskCmd.AddCommand(taskGetCmd)
	taskCmd.AddCommand(taskListCmd)
}

type taskPushInput struct {
	FeatureID      int64  `json:"feature_id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	TargetFile     string `json:"target_file"`
	TargetLine     *int   `json:"target_line"`
	TargetFunction string `json:"target_function"`
}

func runTaskPush(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	var input taskPushInput
	if err := parseJSON(args[0], &input); err != nil {
		outputError(fmt.Errorf("parse JSON: %w", err))
		return nil
	}

	if input.FeatureID == 0 {
		outputError(fmt.Errorf("missing required field: feature_id"))
		return nil
	}
	if input.Title == "" {
		outputError(fmt.Errorf("missing required field: title"))
		return nil
	}
	if input.Content == "" {
		outputError(fmt.Errorf("missing required field: content"))
		return nil
	}

	taskInput := service.TaskCreateInput{
		FeatureID:      input.FeatureID,
		Title:          input.Title,
		Content:        input.Content,
		TargetFile:     input.TargetFile,
		TargetLine:     input.TargetLine,
		TargetFunction: input.TargetFunction,
	}

	taskID, err := service.CreateTask(database, taskInput)
	if err != nil {
		outputError(fmt.Errorf("create task: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"task_id": taskID,
		"title":   input.Title,
		"message": "Task created successfully",
	})

	return nil
}

func runTaskPop(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	result, err := service.PopTask(database)
	if err != nil {
		outputError(fmt.Errorf("pop task: %w", err))
		return nil
	}

	if result.Task == nil {
		outputJSON(map[string]interface{}{
			"success": true,
			"task":    nil,
			"message": "No pending tasks",
		})
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":  true,
		"task":     result.Task,
		"manifest": result.Manifest,
	})

	return nil
}

func runTaskStart(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	taskID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid task ID: %s", args[0]))
		return nil
	}

	if err := service.StartTask(database, taskID); err != nil {
		outputError(fmt.Errorf("start task: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"task_id": taskID,
		"status":  "doing",
		"message": "Task started",
	})

	return nil
}

type taskCompleteInput struct {
	Result string `json:"result"`
	Notes  string `json:"notes"`
}

func runTaskComplete(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	taskID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid task ID: %s", args[0]))
		return nil
	}

	var input taskCompleteInput
	if err := parseJSON(args[1], &input); err != nil {
		outputError(fmt.Errorf("parse JSON: %w", err))
		return nil
	}

	if input.Result == "" {
		outputError(fmt.Errorf("missing required field: result"))
		return nil
	}

	result := input.Result
	if input.Notes != "" {
		result = result + "\n" + input.Notes
	}

	if err := service.CompleteTask(database, taskID, result); err != nil {
		outputError(fmt.Errorf("complete task: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"task_id": taskID,
		"status":  "done",
		"message": "Task completed",
	})

	return nil
}

type taskFailInput struct {
	Error   string `json:"error"`
	Details string `json:"details"`
}

func runTaskFail(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	taskID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid task ID: %s", args[0]))
		return nil
	}

	var input taskFailInput
	if err := parseJSON(args[1], &input); err != nil {
		outputError(fmt.Errorf("parse JSON: %w", err))
		return nil
	}

	if input.Error == "" {
		outputError(fmt.Errorf("missing required field: error"))
		return nil
	}

	errMsg := input.Error
	if input.Details != "" {
		errMsg = errMsg + "\n" + input.Details
	}

	if err := service.FailTask(database, taskID, errMsg); err != nil {
		outputError(fmt.Errorf("fail task: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"task_id": taskID,
		"status":  "failed",
		"message": "Task failed",
	})

	return nil
}

func runTaskStatus(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	status, err := service.GetTaskStatus(database)
	if err != nil {
		outputError(fmt.Errorf("get task status: %w", err))
		return nil
	}

	// Get current state
	states, _ := service.GetAllStates(database)

	outputJSON(map[string]interface{}{
		"success":  true,
		"summary":  status,
		"state":    states,
		"progress": fmt.Sprintf("%.1f%%", status.Progress),
	})

	return nil
}

func runTaskGet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	taskID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid task ID: %s", args[0]))
		return nil
	}

	task, err := service.GetTask(database, taskID)
	if err != nil {
		outputError(fmt.Errorf("get task: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"task":    task,
	})

	return nil
}

func runTaskList(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// If feature_id is provided, filter by feature
	if len(args) > 0 {
		featureID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
			return nil
		}

		tasks, err := service.ListTasksByFeature(database, featureID)
		if err != nil {
			outputError(fmt.Errorf("list tasks: %w", err))
			return nil
		}

		outputJSON(map[string]interface{}{
			"success": true,
			"tasks":   tasks,
			"total":   len(tasks),
		})
		return nil
	}

	// Otherwise list all tasks
	tasks, err := service.ListAllTasks(database)
	if err != nil {
		outputError(fmt.Errorf("list tasks: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"tasks":   tasks,
		"total":   len(tasks),
	})

	return nil
}
