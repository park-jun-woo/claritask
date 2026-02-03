package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"parkjunwoo.com/claritask/internal/service"
)

var initCmd = &cobra.Command{
	Use:   "init <project-id>",
	Short: "Initialize a new project with LLM collaboration",
	Long: `Initialize a new project. This command:
1. Creates .claritask/db database
2. Analyzes project files using LLM
3. Generates tech/design configuration
4. Creates specs document with feedback loop

Use --skip-analysis --skip-specs for quick initialization without LLM.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringP("name", "n", "", "Project name (default: project-id)")
	initCmd.Flags().StringP("description", "d", "", "Project description")
	initCmd.Flags().Bool("skip-analysis", false, "Skip context analysis")
	initCmd.Flags().Bool("skip-specs", false, "Skip specs generation")
	initCmd.Flags().Bool("non-interactive", false, "Non-interactive mode (auto approve)")
	initCmd.Flags().BoolP("interactive", "i", false, "Start interactive requirements gathering with Claude (TTY Handover)")
	initCmd.Flags().Bool("force", false, "Overwrite existing database")
	initCmd.Flags().Bool("resume", false, "Resume interrupted initialization")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Get flags
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	skipAnalysis, _ := cmd.Flags().GetBool("skip-analysis")
	skipSpecs, _ := cmd.Flags().GetBool("skip-specs")
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
	interactive, _ := cmd.Flags().GetBool("interactive")
	force, _ := cmd.Flags().GetBool("force")
	resume, _ := cmd.Flags().GetBool("resume")

	// Handle --resume
	if resume {
		return runInitResume()
	}

	// Project ID is required when not resuming
	if len(args) < 1 {
		outputError(fmt.Errorf("project-id is required"))
		return nil
	}

	projectID := args[0]

	// Validate project ID
	if err := service.ValidateProjectID(projectID); err != nil {
		outputError(err)
		return nil
	}

	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		outputError(fmt.Errorf("get working directory: %w", err))
		return nil
	}

	// Build config
	config := service.InitConfig{
		ProjectID:      projectID,
		Name:           name,
		Description:    description,
		SkipAnalysis:   skipAnalysis,
		SkipSpecs:      skipSpecs,
		NonInteractive: nonInteractive,
		Force:          force,
		WorkDir:        workDir,
	}

	// Set default name
	if config.Name == "" {
		config.Name = projectID
	}

	// Run init process
	result, err := service.RunInit(config)
	if err != nil {
		// Error already printed by service
		return nil
	}

	// Output JSON result for scripting
	if !result.Success {
		outputError(fmt.Errorf(result.Error))
		return nil
	}

	// If interactive mode, start TTY handover for requirements gathering
	if interactive {
		database, err := getDB()
		if err != nil {
			outputError(fmt.Errorf("open database: %w", err))
			return nil
		}
		defer database.Close()

		if err := service.RunInteractiveInit(database, config.ProjectID, config.Name, config.Description); err != nil {
			outputError(fmt.Errorf("interactive init: %w", err))
			return nil
		}
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":    result.Success,
		"project_id": result.ProjectID,
		"db_path":    result.DBPath,
		"specs_path": result.SpecsPath,
	})

	return nil
}

func runInitResume() error {
	workDir, err := os.Getwd()
	if err != nil {
		outputError(fmt.Errorf("get working directory: %w", err))
		return nil
	}

	result, err := service.ResumeInit(workDir)
	if err != nil {
		// Error already printed by service
		return nil
	}

	if !result.Success {
		outputError(fmt.Errorf(result.Error))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":    result.Success,
		"project_id": result.ProjectID,
		"db_path":    result.DBPath,
		"specs_path": result.SpecsPath,
	})

	return nil
}
