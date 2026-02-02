package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"parkjunwoo.com/talos/internal/service"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Project management commands",
}

var projectSetCmd = &cobra.Command{
	Use:   "set '<json>'",
	Short: "Create or update project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectSet,
}

var projectGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get project information",
	RunE:  runProjectGet,
}

var projectPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Start project planning",
	RunE:  runProjectPlan,
}

var projectStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start project execution",
	RunE:  runProjectStart,
}

func init() {
	projectCmd.AddCommand(projectSetCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectPlanCmd)
	projectCmd.AddCommand(projectStartCmd)
}

func runProjectSet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	var input service.ProjectSetInput
	if err := parseJSON(args[0], &input); err != nil {
		outputError(fmt.Errorf("parse JSON: %w", err))
		return nil
	}

	if err := service.SetProjectFull(database, input); err != nil {
		outputError(fmt.Errorf("set project: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"message": "Project updated successfully",
	})

	return nil
}

func runProjectGet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	project, err := service.GetProject(database)
	if err != nil {
		outputError(fmt.Errorf("get project: %w", err))
		return nil
	}

	result := map[string]interface{}{
		"success": true,
		"project": map[string]interface{}{
			"id":          project.ID,
			"name":        project.Name,
			"description": project.Description,
			"status":      project.Status,
			"created_at":  project.CreatedAt,
		},
	}

	// Get context
	if ctx, err := service.GetContext(database); err == nil {
		result["context"] = ctx
	}

	// Get tech
	if tech, err := service.GetTech(database); err == nil {
		result["tech"] = tech
	}

	// Get design
	if design, err := service.GetDesign(database); err == nil {
		result["design"] = design
	}

	outputJSON(result)
	return nil
}

func runProjectPlan(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Check required fields
	required, err := service.CheckRequired(database)
	if err != nil {
		outputError(fmt.Errorf("check required: %w", err))
		return nil
	}

	if !required.Ready {
		outputJSON(map[string]interface{}{
			"success":          false,
			"ready":            false,
			"missing_required": required.MissingRequired,
			"message":          "Please configure required settings before planning",
		})
		return nil
	}

	// Get project info
	project, err := service.GetProject(database)
	if err != nil {
		outputError(fmt.Errorf("get project: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":    true,
		"ready":      true,
		"project_id": project.ID,
		"mode":       "planning",
		"message":    "Project is ready for planning. Use 'talos phase create' to add phases.",
	})

	return nil
}

func runProjectStart(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Check required fields
	required, err := service.CheckRequired(database)
	if err != nil {
		outputError(fmt.Errorf("check required: %w", err))
		return nil
	}

	if !required.Ready {
		outputJSON(map[string]interface{}{
			"success":          false,
			"ready":            false,
			"missing_required": required.MissingRequired,
			"message":          "Please configure required settings before starting",
		})
		return nil
	}

	// Get task status
	taskStatus, err := service.GetTaskStatus(database)
	if err != nil {
		outputError(fmt.Errorf("get task status: %w", err))
		return nil
	}

	if taskStatus.Total == 0 {
		outputJSON(map[string]interface{}{
			"success": false,
			"ready":   false,
			"message": "No tasks found. Use 'talos project plan' and 'talos task push' to add tasks first.",
		})
		return nil
	}

	if taskStatus.Pending == 0 && taskStatus.Doing == 0 {
		outputJSON(map[string]interface{}{
			"success":  true,
			"ready":    true,
			"mode":     "completed",
			"status":   taskStatus,
			"message":  "All tasks are completed!",
			"progress": taskStatus.Progress,
		})
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":  true,
		"ready":    true,
		"mode":     "execution",
		"status":   taskStatus,
		"message":  "Project is ready for execution. Use 'talos task pop' to get the next task.",
		"progress": taskStatus.Progress,
	})

	return nil
}
