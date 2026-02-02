package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"parkjunwoo.com/talos/internal/service"
)

var techCmd = &cobra.Command{
	Use:   "tech",
	Short: "Tech stack management",
}

var techSetCmd = &cobra.Command{
	Use:   "set '<json>'",
	Short: "Set tech stack",
	Args:  cobra.ExactArgs(1),
	RunE:  runTechSet,
}

var techGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get tech stack",
	RunE:  runTechGet,
}

func init() {
	techCmd.AddCommand(techSetCmd)
	techCmd.AddCommand(techGetCmd)
}

func runTechSet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	var data map[string]interface{}
	if err := parseJSON(args[0], &data); err != nil {
		outputError(fmt.Errorf("parse JSON: %w", err))
		return nil
	}

	// Validate required fields
	if err := validateTech(data); err != nil {
		outputError(err)
		return nil
	}

	if err := service.SetTech(database, data); err != nil {
		outputError(fmt.Errorf("set tech: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"message": "Tech updated successfully",
	})

	return nil
}

func runTechGet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	data, err := service.GetTech(database)
	if err != nil {
		outputError(fmt.Errorf("get tech: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"tech":    data,
	})

	return nil
}

func validateTech(data map[string]interface{}) error {
	required := []string{"backend", "frontend", "database"}
	for _, field := range required {
		if _, ok := data[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}
