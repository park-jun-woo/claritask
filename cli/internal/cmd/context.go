package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"parkjunwoo.com/claritask/internal/service"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Context management",
}

var contextSetCmd = &cobra.Command{
	Use:   "set '<json>'",
	Short: "Set project context",
	Args:  cobra.ExactArgs(1),
	RunE:  runContextSet,
}

var contextGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get project context",
	RunE:  runContextGet,
}

func init() {
	contextCmd.AddCommand(contextSetCmd)
	contextCmd.AddCommand(contextGetCmd)
}

func runContextSet(cmd *cobra.Command, args []string) error {
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
	if err := validateContext(data); err != nil {
		outputError(err)
		return nil
	}

	if err := service.SetContext(database, data); err != nil {
		outputError(fmt.Errorf("set context: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"message": "Context updated successfully",
	})

	return nil
}

func runContextGet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	data, err := service.GetContext(database)
	if err != nil {
		outputError(fmt.Errorf("get context: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"context": data,
	})

	return nil
}

func validateContext(data map[string]interface{}) error {
	required := []string{"project_name", "description"}
	for _, field := range required {
		if _, ok := data[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}
