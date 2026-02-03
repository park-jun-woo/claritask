package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"parkjunwoo.com/claritask/internal/service"
)

var featureCmd = &cobra.Command{
	Use:   "feature",
	Short: "Feature management commands",
}

var featureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all features",
	RunE:  runFeatureList,
}

var featureAddCmd = &cobra.Command{
	Use:   "add '<json>'",
	Short: "Add a new feature",
	Args:  cobra.ExactArgs(1),
	RunE:  runFeatureAdd,
}

var featureGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get feature details",
	Args:  cobra.ExactArgs(1),
	RunE:  runFeatureGet,
}

var featureSpecCmd = &cobra.Command{
	Use:   "spec <id> '<spec>'",
	Short: "Set feature spec",
	Args:  cobra.ExactArgs(2),
	RunE:  runFeatureSpec,
}

var featureStartCmd = &cobra.Command{
	Use:   "start <id>",
	Short: "Start feature execution",
	Args:  cobra.ExactArgs(1),
	RunE:  runFeatureStart,
}

var featureTasksCmd = &cobra.Command{
	Use:   "tasks <id>",
	Short: "List or generate tasks for a feature",
	Args:  cobra.ExactArgs(1),
	RunE:  runFeatureTasks,
}

var featureCreateCmd = &cobra.Command{
	Use:   "create '<json>'",
	Short: "Create feature with FDL and tasks in one step",
	Args:  cobra.ExactArgs(1),
	RunE:  runFeatureCreate,
}

func init() {
	featureCmd.AddCommand(featureListCmd)
	featureCmd.AddCommand(featureAddCmd)
	featureCmd.AddCommand(featureGetCmd)
	featureCmd.AddCommand(featureSpecCmd)
	featureCmd.AddCommand(featureStartCmd)
	featureCmd.AddCommand(featureTasksCmd)
	featureCmd.AddCommand(featureCreateCmd)

	// feature tasks flags
	featureTasksCmd.Flags().Bool("generate", false, "Generate tasks using LLM (when no FDL)")
}

type featureAddInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type featureCreateInput struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	FDL              string `json:"fdl,omitempty"`
	GenerateTasks    bool   `json:"generate_tasks"`
	GenerateSkeleton bool   `json:"generate_skeleton"`
}

func runFeatureList(cmd *cobra.Command, args []string) error {
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

	features, err := service.ListFeaturesWithStats(database, project.ID)
	if err != nil {
		outputError(fmt.Errorf("list features: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":  true,
		"features": features,
		"total":    len(features),
	})

	return nil
}

func runFeatureAdd(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	var input featureAddInput
	if err := parseJSON(args[0], &input); err != nil {
		outputError(fmt.Errorf("parse JSON: %w", err))
		return nil
	}

	if input.Name == "" {
		outputError(fmt.Errorf("missing required field: name"))
		return nil
	}

	if input.Description == "" {
		outputError(fmt.Errorf("missing required field: description"))
		return nil
	}

	project, err := service.GetProject(database)
	if err != nil {
		outputError(fmt.Errorf("get project: %w", err))
		return nil
	}

	result, err := service.CreateFeature(database, project.ID, input.Name, input.Description)
	if err != nil {
		outputError(fmt.Errorf("create feature: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":    true,
		"feature_id": result.ID,
		"name":       input.Name,
		"file_path":  result.FilePath,
		"message":    "Feature created successfully",
	})

	return nil
}

func runFeatureGet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	featureID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
		return nil
	}

	feature, err := service.GetFeature(database, featureID)
	if err != nil {
		outputError(fmt.Errorf("get feature: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"feature": map[string]interface{}{
			"id":                 feature.ID,
			"name":               feature.Name,
			"description":        feature.Description,
			"spec":               feature.Spec,
			"status":             feature.Status,
			"fdl":                feature.FDL,
			"fdl_hash":           feature.FDLHash,
			"skeleton_generated": feature.SkeletonGenerated,
			"file_path":          feature.FilePath,
			"content":            feature.Content,
			"content_hash":       feature.ContentHash,
			"created_at":         feature.CreatedAt,
		},
	})

	return nil
}

func runFeatureSpec(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	featureID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
		return nil
	}

	spec := args[1]

	if err := service.SetFeatureSpec(database, featureID, spec); err != nil {
		outputError(fmt.Errorf("set feature spec: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":    true,
		"feature_id": featureID,
		"message":    "Feature spec updated successfully",
	})

	return nil
}

func runFeatureStart(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	featureID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
		return nil
	}

	feature, err := service.GetFeature(database, featureID)
	if err != nil {
		outputError(fmt.Errorf("get feature: %w", err))
		return nil
	}

	// Start feature if pending
	if feature.Status == "pending" {
		if err := service.StartFeature(database, featureID); err != nil {
			outputError(fmt.Errorf("start feature: %w", err))
			return nil
		}
	}

	// Get task count for this feature
	var pendingTasks int
	row := database.QueryRow(`SELECT COUNT(*) FROM tasks WHERE feature_id = ? AND status = 'pending'`, featureID)
	row.Scan(&pendingTasks)

	outputJSON(map[string]interface{}{
		"success":       true,
		"feature_id":    featureID,
		"name":          feature.Name,
		"mode":          "execution",
		"pending_tasks": pendingTasks,
		"message":       "Feature execution started",
	})

	return nil
}

func runFeatureTasks(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	featureID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
		return nil
	}

	generate, _ := cmd.Flags().GetBool("generate")

	// Get feature
	feature, err := service.GetFeature(database, featureID)
	if err != nil {
		outputError(fmt.Errorf("get feature: %w", err))
		return nil
	}

	// If feature has FDL, use fdl tasks logic
	if feature.FDL != "" && !generate {
		// Get tasks for this feature
		tasks, err := service.ListTasksByFeature(database, featureID)
		if err != nil {
			outputError(fmt.Errorf("list tasks: %w", err))
			return nil
		}

		var taskList []map[string]interface{}
		for _, t := range tasks {
			taskList = append(taskList, map[string]interface{}{
				"id":     t.ID,
				"title":  t.Title,
				"status": t.Status,
			})
		}

		outputJSON(map[string]interface{}{
			"success":      true,
			"feature_id":   featureID,
			"feature_name": feature.Name,
			"tasks":        taskList,
			"total":        len(taskList),
		})
		return nil
	}

	// Generate tasks using LLM (prepare prompt)
	context := map[string]interface{}{
		"feature_id":   featureID,
		"feature_name": feature.Name,
		"spec":         feature.Spec,
	}

	// Get tech stack for context
	tech, _ := service.GetTech(database)
	if tech != nil {
		context["tech"] = tech
	}

	prompt := fmt.Sprintf(`Analyze the feature "%s" and generate implementation tasks.

Feature Name: %s
Spec: %s

Generate a list of tasks needed to implement this feature. For each task provide:
1. title: Short descriptive name (snake_case)
2. content: Detailed implementation instructions
3. level: "leaf" for implementation tasks
4. dependencies: List of other task titles this depends on

Return JSON array:
[{"title": "task_name", "content": "...", "level": "leaf", "dependencies": [...]}]
`, feature.Name, feature.Name, feature.Spec)

	outputJSON(map[string]interface{}{
		"success":      true,
		"feature_id":   featureID,
		"feature_name": feature.Name,
		"context":      context,
		"prompt":       prompt,
		"instructions": "Use the prompt to generate tasks, then run 'clari task push' for each task",
	})

	return nil
}

func runFeatureCreate(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	var input featureCreateInput
	if err := parseJSON(args[0], &input); err != nil {
		outputError(fmt.Errorf("parse JSON: %w", err))
		return nil
	}

	if input.Name == "" {
		outputError(fmt.Errorf("missing required field: name"))
		return nil
	}

	if input.Description == "" {
		outputError(fmt.Errorf("missing required field: description"))
		return nil
	}

	project, err := service.GetProject(database)
	if err != nil {
		outputError(fmt.Errorf("get project: %w", err))
		return nil
	}

	// Step 1: Create feature with md file
	result, err := service.CreateFeature(database, project.ID, input.Name, input.Description)
	if err != nil {
		outputError(fmt.Errorf("create feature: %w", err))
		return nil
	}

	response := map[string]interface{}{
		"success":    true,
		"feature_id": result.ID,
		"name":       input.Name,
		"file_path":  result.FilePath,
	}

	// Step 2: If FDL provided, register and validate it
	if input.FDL != "" {
		if err := service.SetFeatureFDL(database, result.ID, input.FDL); err != nil {
			response["fdl_valid"] = false
			response["fdl_errors"] = []string{err.Error()}
			response["message"] = "Feature created but FDL registration failed"
			outputJSON(response)
			return nil
		}

		// Validate FDL
		fdlValid, fdlErrors := service.ValidateFDLContent(input.FDL)
		response["fdl_hash"] = service.CalculateFDLHash(input.FDL)
		response["fdl_valid"] = fdlValid
		if !fdlValid {
			response["fdl_errors"] = fdlErrors
			response["message"] = "Feature created but FDL validation failed"
			outputJSON(response)
			return nil
		}

		// Step 3: Generate tasks if requested and FDL is valid
		if input.GenerateTasks && fdlValid {
			tasksCreated, edgesCreated, err := service.GenerateTasksFromFDL(database, result.ID)
			if err != nil {
				response["tasks_created"] = 0
				response["edges_created"] = 0
				response["message"] = fmt.Sprintf("Feature created with FDL but task generation failed: %v", err)
			} else {
				response["tasks_created"] = tasksCreated
				response["edges_created"] = edgesCreated
			}
		}

		// Step 4: Generate skeleton if requested
		if input.GenerateSkeleton && fdlValid {
			skeletonFiles, err := service.GenerateSkeletonFromFDL(database, result.ID)
			if err != nil {
				response["skeleton_files"] = []string{}
			} else {
				response["skeleton_files"] = skeletonFiles
			}
		}
	}

	if _, ok := response["message"]; !ok {
		tasksCreated, _ := response["tasks_created"].(int)
		if tasksCreated > 0 {
			response["message"] = fmt.Sprintf("Feature created with FDL and %d tasks", tasksCreated)
		} else if input.FDL != "" {
			response["message"] = "Feature created with FDL"
		} else {
			response["message"] = "Feature created successfully"
		}
	}

	outputJSON(response)
	return nil
}
