package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"parkjunwoo.com/claritask/internal/service"
)

var fdlCmd = &cobra.Command{
	Use:   "fdl",
	Short: "FDL (Feature Definition Language) management",
}

var fdlCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create FDL template file",
	Args:  cobra.ExactArgs(1),
	RunE:  runFDLCreate,
}

var fdlRegisterCmd = &cobra.Command{
	Use:   "register <file>",
	Short: "Register FDL file as a feature",
	Args:  cobra.ExactArgs(1),
	RunE:  runFDLRegister,
}

var fdlValidateCmd = &cobra.Command{
	Use:   "validate <feature_id>",
	Short: "Validate FDL for a feature",
	Args:  cobra.ExactArgs(1),
	RunE:  runFDLValidate,
}

var fdlShowCmd = &cobra.Command{
	Use:   "show <feature_id>",
	Short: "Show FDL content for a feature",
	Args:  cobra.ExactArgs(1),
	RunE:  runFDLShow,
}

var fdlSkeletonCmd = &cobra.Command{
	Use:   "skeleton <feature_id>",
	Short: "Generate skeleton files from FDL",
	Args:  cobra.ExactArgs(1),
	RunE:  runFDLSkeleton,
}

var fdlTasksCmd = &cobra.Command{
	Use:   "tasks <feature_id>",
	Short: "Generate tasks from FDL",
	Args:  cobra.ExactArgs(1),
	RunE:  runFDLTasks,
}

var fdlVerifyCmd = &cobra.Command{
	Use:   "verify <feature_id>",
	Short: "Verify implementation matches FDL",
	Args:  cobra.ExactArgs(1),
	RunE:  runFDLVerify,
}

var fdlDiffCmd = &cobra.Command{
	Use:   "diff <feature_id>",
	Short: "Show differences between FDL and actual code",
	Args:  cobra.ExactArgs(1),
	RunE:  runFDLDiff,
}

func init() {
	fdlCmd.AddCommand(fdlCreateCmd)
	fdlCmd.AddCommand(fdlRegisterCmd)
	fdlCmd.AddCommand(fdlValidateCmd)
	fdlCmd.AddCommand(fdlShowCmd)
	fdlCmd.AddCommand(fdlSkeletonCmd)
	fdlCmd.AddCommand(fdlTasksCmd)
	fdlCmd.AddCommand(fdlVerifyCmd)
	fdlCmd.AddCommand(fdlDiffCmd)

	// fdl skeleton flags
	fdlSkeletonCmd.Flags().Bool("dry-run", false, "Show files to be generated without creating them")
	fdlSkeletonCmd.Flags().Bool("force", false, "Overwrite existing skeleton files")
}

func runFDLCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Create features directory if not exists
	featuresDir := "features"
	if err := os.MkdirAll(featuresDir, 0755); err != nil {
		outputError(fmt.Errorf("create features directory: %w", err))
		return nil
	}

	// Generate template
	template := service.GenerateFDLTemplate(name)

	// Write file
	filePath := filepath.Join(featuresDir, name+".fdl.yaml")
	if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
		outputError(fmt.Errorf("write FDL file: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"file":    filePath,
		"message": "FDL template created. Edit the file and run 'clari fdl register'",
	})

	return nil
}

func runFDLRegister(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Parse FDL file
	spec, err := service.ParseFDLFile(filePath)
	if err != nil {
		outputError(fmt.Errorf("parse FDL file: %w", err))
		return nil
	}

	// Validate FDL
	if err := service.ValidateFDL(spec); err != nil {
		outputJSON(map[string]interface{}{
			"success": false,
			"error":   "FDL validation failed",
			"details": []string{err.Error()},
		})
		return nil
	}

	// Read file content for hash
	content, err := os.ReadFile(filePath)
	if err != nil {
		outputError(fmt.Errorf("read FDL file: %w", err))
		return nil
	}
	fdlContent := string(content)
	fdlHash := service.CalculateFDLHashFromSpec(fdlContent)

	// Get project
	project, err := service.GetProject(database)
	if err != nil {
		outputError(fmt.Errorf("get project: %w", err))
		return nil
	}

	// Create feature
	result, err := service.CreateFeature(database, project.ID, spec.Feature, spec.Description)
	if err != nil {
		outputError(fmt.Errorf("create feature: %w", err))
		return nil
	}

	// Set FDL
	if err := service.SetFeatureFDL(database, result.ID, fdlContent); err != nil {
		outputError(fmt.Errorf("set feature FDL: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":      true,
		"feature_id":   result.ID,
		"feature_name": spec.Feature,
		"file_path":    result.FilePath,
		"fdl_hash":     fdlHash,
		"message":      "FDL registered successfully",
	})

	return nil
}

func runFDLValidate(cmd *cobra.Command, args []string) error {
	featureID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
		return nil
	}

	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Get FDL
	fdl, err := service.GetFeatureFDL(database, featureID)
	if err != nil {
		outputError(fmt.Errorf("get feature FDL: %w", err))
		return nil
	}

	// Parse FDL
	spec, err := service.ParseFDL(fdl)
	if err != nil {
		outputJSON(map[string]interface{}{
			"success":    true,
			"feature_id": featureID,
			"valid":      false,
			"errors":     []string{err.Error()},
		})
		return nil
	}

	// Validate FDL
	if err := service.ValidateFDL(spec); err != nil {
		outputJSON(map[string]interface{}{
			"success":    true,
			"feature_id": featureID,
			"valid":      false,
			"errors":     []string{err.Error()},
		})
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":    true,
		"feature_id": featureID,
		"valid":      true,
		"message":    "FDL is valid",
	})

	return nil
}

func runFDLShow(cmd *cobra.Command, args []string) error {
	featureID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
		return nil
	}

	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	feature, err := service.GetFeature(database, featureID)
	if err != nil {
		outputError(fmt.Errorf("get feature: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":            true,
		"feature_id":         feature.ID,
		"feature_name":       feature.Name,
		"fdl":                feature.FDL,
		"fdl_hash":           feature.FDLHash,
		"skeleton_generated": feature.SkeletonGenerated,
	})

	return nil
}

func runFDLSkeleton(cmd *cobra.Command, args []string) error {
	featureID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
		return nil
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")

	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Get feature
	feature, err := service.GetFeature(database, featureID)
	if err != nil {
		outputError(fmt.Errorf("get feature: %w", err))
		return nil
	}

	// Check if skeleton already generated
	if feature.SkeletonGenerated && !force && !dryRun {
		outputJSON(map[string]interface{}{
			"success":    false,
			"feature_id": featureID,
			"error":      "Skeleton already generated. Use --force to regenerate.",
		})
		return nil
	}

	// Generate skeletons using Go implementation
	result, err := service.GenerateSkeletons(database, featureID, dryRun)
	if err != nil {
		outputError(fmt.Errorf("generate skeletons: %w", err))
		return nil
	}

	// Build file list for response
	var generatedFiles []map[string]interface{}
	for _, f := range result.Files {
		generatedFiles = append(generatedFiles, map[string]interface{}{
			"path":  f.Path,
			"layer": f.Layer,
		})
	}

	if dryRun {
		outputJSON(map[string]interface{}{
			"success":         true,
			"feature_id":      featureID,
			"mode":            "dry-run",
			"generated_files": generatedFiles,
			"total":           len(generatedFiles),
			"message":         "Files that would be generated (dry-run)",
		})
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":         true,
		"feature_id":      featureID,
		"generated_files": generatedFiles,
		"total":           len(generatedFiles),
		"errors":          result.Errors,
		"message":         "Skeletons generated successfully",
	})

	return nil
}

func runFDLTasks(cmd *cobra.Command, args []string) error {
	featureID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
		return nil
	}

	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Get feature
	feature, err := service.GetFeature(database, featureID)
	if err != nil {
		outputError(fmt.Errorf("get feature: %w", err))
		return nil
	}

	// Parse FDL
	spec, err := service.ParseFDL(feature.FDL)
	if err != nil {
		outputError(fmt.Errorf("parse FDL: %w", err))
		return nil
	}

	// Get tech stack
	tech, _ := service.GetTech(database)
	if tech == nil {
		tech = make(map[string]interface{})
	}

	// Extract task mappings
	mappings, err := service.ExtractTaskMappings(spec, tech)
	if err != nil {
		outputError(fmt.Errorf("extract task mappings: %w", err))
		return nil
	}

	// Create tasks
	var tasksCreated []map[string]interface{}
	taskIDMap := make(map[string]int64)

	for _, m := range mappings {
		taskInput := service.TaskCreateInput{
			FeatureID:      featureID,
			Title:          m.Title,
			Content:        m.Content,
			TargetFile:     m.TargetFile,
			TargetFunction: m.TargetFunction,
		}

		taskID, err := service.CreateTask(database, taskInput)
		if err != nil {
			outputError(fmt.Errorf("create task: %w", err))
			return nil
		}

		taskIDMap[m.Title] = taskID

		tasksCreated = append(tasksCreated, map[string]interface{}{
			"id":              taskID,
			"title":           m.Title,
			"target_file":     m.TargetFile,
			"target_function": m.TargetFunction,
		})
	}

	// Create edges based on dependencies
	var edgesCreated []map[string]interface{}
	for _, m := range mappings {
		fromID := taskIDMap[m.Title]
		for _, depTitle := range m.Dependencies {
			if toID, ok := taskIDMap[depTitle]; ok {
				if err := service.AddTaskEdge(database, fromID, toID); err == nil {
					edgesCreated = append(edgesCreated, map[string]interface{}{
						"from": fromID,
						"to":   toID,
					})
				}
			}
		}
	}

	outputJSON(map[string]interface{}{
		"success":       true,
		"feature_id":    featureID,
		"tasks_created": tasksCreated,
		"edges_created": edgesCreated,
		"total_tasks":   len(tasksCreated),
		"total_edges":   len(edgesCreated),
		"message":       "Tasks generated from FDL",
	})

	return nil
}

func runFDLVerify(cmd *cobra.Command, args []string) error {
	featureID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
		return nil
	}

	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Verify FDL implementation
	result, err := service.VerifyFDLImplementation(database, featureID)
	if err != nil {
		outputError(fmt.Errorf("verify FDL implementation: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":            true,
		"feature_id":         featureID,
		"valid":              result.Valid,
		"errors":             result.Errors,
		"warnings":           result.Warnings,
		"functions_missing":  result.FunctionsMissing,
		"functions_extra":    result.FunctionsExtra,
		"files_missing":      result.FilesMissing,
		"signature_mismatch": result.SignatureMismatch,
		"models_missing":     result.ModelsMissing,
		"apis_missing":       result.APIsMissing,
	})

	return nil
}

func runFDLDiff(cmd *cobra.Command, args []string) error {
	featureID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid feature ID: %s", args[0]))
		return nil
	}

	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Get FDL diff
	result, err := service.DiffFDLImplementation(database, featureID)
	if err != nil {
		outputError(fmt.Errorf("diff FDL implementation: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":       true,
		"feature_id":    result.FeatureID,
		"feature_name":  result.FeatureName,
		"differences":   result.Differences,
		"total_changes": result.TotalChanges,
	})

	return nil
}
