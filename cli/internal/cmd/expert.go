package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"parkjunwoo.com/claritask/internal/service"
)

var expertCmd = &cobra.Command{
	Use:   "expert",
	Short: "Expert management commands",
}

var expertAddCmd = &cobra.Command{
	Use:   "add <expert-id>",
	Short: "Create a new expert",
	Args:  cobra.ExactArgs(1),
	RunE:  runExpertAdd,
}

var expertListCmd = &cobra.Command{
	Use:   "list",
	Short: "List experts",
	RunE:  runExpertList,
}

var expertGetCmd = &cobra.Command{
	Use:   "get <expert-id>",
	Short: "Get expert details",
	Args:  cobra.ExactArgs(1),
	RunE:  runExpertGet,
}

var expertEditCmd = &cobra.Command{
	Use:   "edit <expert-id>",
	Short: "Edit expert file",
	Args:  cobra.ExactArgs(1),
	RunE:  runExpertEdit,
}

var expertRemoveCmd = &cobra.Command{
	Use:   "remove <expert-id>",
	Short: "Remove an expert",
	Args:  cobra.ExactArgs(1),
	RunE:  runExpertRemove,
}

var expertAssignCmd = &cobra.Command{
	Use:   "assign <expert-id>",
	Short: "Assign expert to project",
	Args:  cobra.ExactArgs(1),
	RunE:  runExpertAssign,
}

var expertUnassignCmd = &cobra.Command{
	Use:   "unassign <expert-id>",
	Short: "Unassign expert from project",
	Args:  cobra.ExactArgs(1),
	RunE:  runExpertUnassign,
}

var (
	expertListAssigned  bool
	expertListAvailable bool
	expertRemoveForce   bool
	expertAssignProject string
	expertAssignFeature int64
	expertAddOpenEditor bool
)

func init() {
	expertCmd.AddCommand(expertAddCmd)
	expertCmd.AddCommand(expertListCmd)
	expertCmd.AddCommand(expertGetCmd)
	expertCmd.AddCommand(expertEditCmd)
	expertCmd.AddCommand(expertRemoveCmd)
	expertCmd.AddCommand(expertAssignCmd)
	expertCmd.AddCommand(expertUnassignCmd)

	expertAddCmd.Flags().BoolVar(&expertAddOpenEditor, "edit", false, "Open editor after creation")

	expertListCmd.Flags().BoolVar(&expertListAssigned, "assigned", false, "List only assigned experts")
	expertListCmd.Flags().BoolVar(&expertListAvailable, "available", false, "List only available experts")

	expertRemoveCmd.Flags().BoolVar(&expertRemoveForce, "force", false, "Force remove even if assigned")

	expertAssignCmd.Flags().StringVar(&expertAssignProject, "project", "", "Project ID (default: current project)")
	expertAssignCmd.Flags().Int64Var(&expertAssignFeature, "feature", 0, "Feature ID for feature-level assignment")
	expertUnassignCmd.Flags().StringVar(&expertAssignProject, "project", "", "Project ID (default: current project)")
	expertUnassignCmd.Flags().Int64Var(&expertAssignFeature, "feature", 0, "Feature ID for feature-level unassignment")
}

func runExpertAdd(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	expertID := args[0]
	expert, err := service.AddExpert(database, expertID)
	if err != nil {
		outputError(err)
		return nil
	}

	// Open editor if --edit flag is set
	if expertAddOpenEditor {
		openEditor(expert.Path)
	}

	outputJSON(map[string]interface{}{
		"success":   true,
		"expert_id": expert.ID,
		"path":      expert.Path,
		"message":   "Expert created. Edit the file to define the expert.",
	})

	return nil
}

// openEditor opens the specified file in the system editor
func openEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vi"
		}
	}

	execCmd := exec.Command(editor, filePath)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

func runExpertList(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	filter := "all"
	if expertListAssigned {
		filter = "assigned"
	} else if expertListAvailable {
		filter = "available"
	}

	experts, err := service.ListExperts(database, filter)
	if err != nil {
		outputError(fmt.Errorf("list experts: %w", err))
		return nil
	}

	var expertList []map[string]interface{}
	for _, e := range experts {
		expertList = append(expertList, map[string]interface{}{
			"id":        e.ID,
			"name":      e.Name,
			"domain":    e.Domain,
			"assigned":  e.Assigned,
		})
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"experts": expertList,
		"total":   len(expertList),
	})

	return nil
}

func runExpertGet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	expertID := args[0]
	expert, err := service.GetExpert(database, expertID)
	if err != nil {
		outputError(err)
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"expert": map[string]interface{}{
			"id":        expert.ID,
			"name":      expert.Name,
			"version":   expert.Version,
			"domain":    expert.Domain,
			"language":  expert.Language,
			"framework": expert.Framework,
			"path":      expert.Path,
			"assigned":  expert.Assigned,
		},
	})

	return nil
}

func runExpertEdit(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	expertID := args[0]
	expert, err := service.GetExpert(database, expertID)
	if err != nil {
		outputError(err)
		return nil
	}

	if err := openEditor(expert.Path); err != nil {
		outputError(fmt.Errorf("run editor: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":   true,
		"expert_id": expertID,
		"path":      expert.Path,
		"message":   "Editor closed",
	})

	return nil
}

func runExpertRemove(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	expertID := args[0]
	if err := service.RemoveExpert(database, expertID, expertRemoveForce); err != nil {
		outputError(err)
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":   true,
		"expert_id": expertID,
		"message":   "Expert removed successfully",
	})

	return nil
}

func runExpertAssign(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	expertID := args[0]

	// Feature-level assignment
	if expertAssignFeature > 0 {
		if err := service.AssignExpertToFeature(database, expertID, expertAssignFeature); err != nil {
			outputError(err)
			return nil
		}

		outputJSON(map[string]interface{}{
			"success":    true,
			"expert_id":  expertID,
			"feature_id": expertAssignFeature,
			"message":    "Expert assigned to feature",
		})
		return nil
	}

	// Project-level assignment
	projectID := expertAssignProject
	if projectID == "" {
		project, err := service.GetProject(database)
		if err != nil || project == nil {
			outputError(fmt.Errorf("no current project. Use --project to specify"))
			return nil
		}
		projectID = project.ID
	}

	if err := service.AssignExpert(database, projectID, expertID); err != nil {
		outputError(err)
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":    true,
		"expert_id":  expertID,
		"project_id": projectID,
		"message":    "Expert assigned to project",
	})

	return nil
}

func runExpertUnassign(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	expertID := args[0]

	// Feature-level unassignment
	if expertAssignFeature > 0 {
		if err := service.UnassignExpertFromFeature(database, expertID, expertAssignFeature); err != nil {
			outputError(err)
			return nil
		}

		outputJSON(map[string]interface{}{
			"success":    true,
			"expert_id":  expertID,
			"feature_id": expertAssignFeature,
			"message":    "Expert unassigned from feature",
		})
		return nil
	}

	// Project-level unassignment
	projectID := expertAssignProject
	if projectID == "" {
		project, err := service.GetProject(database)
		if err != nil || project == nil {
			outputError(fmt.Errorf("no current project. Use --project to specify"))
			return nil
		}
		projectID = project.ID
	}

	if err := service.UnassignExpert(database, projectID, expertID); err != nil {
		outputError(err)
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":    true,
		"expert_id":  expertID,
		"project_id": projectID,
		"message":    "Expert unassigned from project",
	})

	return nil
}
