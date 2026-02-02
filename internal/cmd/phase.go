package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"parkjunwoo.com/talos/internal/service"
)

var phaseCmd = &cobra.Command{
	Use:   "phase",
	Short: "Phase management commands",
}

var phaseCreateCmd = &cobra.Command{
	Use:   "create '<json>'",
	Short: "Create a new phase",
	Args:  cobra.ExactArgs(1),
	RunE:  runPhaseCreate,
}

var phaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all phases",
	RunE:  runPhaseList,
}

var phasePlanCmd = &cobra.Command{
	Use:   "plan <phase_id>",
	Short: "Start planning for a phase",
	Args:  cobra.ExactArgs(1),
	RunE:  runPhasePlan,
}

var phaseStartCmd = &cobra.Command{
	Use:   "start <phase_id>",
	Short: "Start execution for a phase",
	Args:  cobra.ExactArgs(1),
	RunE:  runPhaseStart,
}

func init() {
	phaseCmd.AddCommand(phaseCreateCmd)
	phaseCmd.AddCommand(phaseListCmd)
	phaseCmd.AddCommand(phasePlanCmd)
	phaseCmd.AddCommand(phaseStartCmd)
}

type phaseCreateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	OrderNum    int    `json:"order_num"`
}

func runPhaseCreate(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	var input phaseCreateInput
	if err := parseJSON(args[0], &input); err != nil {
		outputError(fmt.Errorf("parse JSON: %w", err))
		return nil
	}

	if input.Name == "" {
		outputError(fmt.Errorf("missing required field: name"))
		return nil
	}

	// Get project
	project, err := service.GetProject(database)
	if err != nil {
		outputError(fmt.Errorf("get project: %w", err))
		return nil
	}

	phaseInput := service.PhaseCreateInput{
		ProjectID:   project.ID,
		Name:        input.Name,
		Description: input.Description,
		OrderNum:    input.OrderNum,
	}

	phaseID, err := service.CreatePhase(database, phaseInput)
	if err != nil {
		outputError(fmt.Errorf("create phase: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":  true,
		"phase_id": phaseID,
		"name":     input.Name,
		"message":  "Phase created successfully",
	})

	return nil
}

func runPhaseList(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Get project
	project, err := service.GetProject(database)
	if err != nil {
		outputError(fmt.Errorf("get project: %w", err))
		return nil
	}

	phases, err := service.ListPhases(database, project.ID)
	if err != nil {
		outputError(fmt.Errorf("list phases: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"phases":  phases,
		"total":   len(phases),
	})

	return nil
}

func runPhasePlan(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	phaseID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid phase ID: %s", args[0]))
		return nil
	}

	phase, err := service.GetPhase(database, phaseID)
	if err != nil {
		outputError(fmt.Errorf("get phase: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":  true,
		"phase_id": phaseID,
		"name":     phase.Name,
		"status":   phase.Status,
		"mode":     "planning",
		"message":  fmt.Sprintf("Phase '%s' is ready for planning. Use 'talos task push' to add tasks.", phase.Name),
	})

	return nil
}

func runPhaseStart(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	phaseID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid phase ID: %s", args[0]))
		return nil
	}

	phase, err := service.GetPhase(database, phaseID)
	if err != nil {
		outputError(fmt.Errorf("get phase: %w", err))
		return nil
	}

	// Get tasks for this phase
	tasks, err := service.ListTasks(database, phaseID)
	if err != nil {
		outputError(fmt.Errorf("list tasks: %w", err))
		return nil
	}

	if len(tasks) == 0 {
		outputJSON(map[string]interface{}{
			"success":  false,
			"phase_id": phaseID,
			"name":     phase.Name,
			"message":  "No tasks found in this phase. Use 'talos task push' to add tasks first.",
		})
		return nil
	}

	// Count pending tasks
	pending := 0
	for _, t := range tasks {
		if t.Status == "pending" {
			pending++
		}
	}

	if pending == 0 {
		outputJSON(map[string]interface{}{
			"success":  true,
			"phase_id": phaseID,
			"name":     phase.Name,
			"status":   phase.Status,
			"mode":     "completed",
			"message":  "No pending tasks in this phase.",
		})
		return nil
	}

	// Start phase if pending
	if phase.Status == "pending" {
		if err := service.StartPhase(database, phaseID); err != nil {
			outputError(fmt.Errorf("start phase: %w", err))
			return nil
		}
	}

	outputJSON(map[string]interface{}{
		"success":       true,
		"phase_id":      phaseID,
		"name":          phase.Name,
		"mode":          "execution",
		"pending_tasks": pending,
		"total_tasks":   len(tasks),
		"message":       fmt.Sprintf("Phase '%s' started. Use 'talos task pop' to get the next task.", phase.Name),
	})

	return nil
}
