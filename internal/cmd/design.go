package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"parkjunwoo.com/talos/internal/service"
)

var designCmd = &cobra.Command{
	Use:   "design",
	Short: "Design decisions management",
}

var designSetCmd = &cobra.Command{
	Use:   "set '<json>'",
	Short: "Set design decisions",
	Args:  cobra.ExactArgs(1),
	RunE:  runDesignSet,
}

var designGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get design decisions",
	RunE:  runDesignGet,
}

func init() {
	designCmd.AddCommand(designSetCmd)
	designCmd.AddCommand(designGetCmd)
}

func runDesignSet(cmd *cobra.Command, args []string) error {
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
	if err := validateDesign(data); err != nil {
		outputError(err)
		return nil
	}

	if err := service.SetDesign(database, data); err != nil {
		outputError(fmt.Errorf("set design: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"message": "Design updated successfully",
	})

	return nil
}

func runDesignGet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	data, err := service.GetDesign(database)
	if err != nil {
		outputError(fmt.Errorf("get design: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"design":  data,
	})

	return nil
}

func validateDesign(data map[string]interface{}) error {
	required := []string{"architecture", "auth_method", "api_style"}
	for _, field := range required {
		if _, ok := data[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}
