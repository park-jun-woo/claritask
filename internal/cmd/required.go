package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"parkjunwoo.com/talos/internal/service"
)

var requiredCmd = &cobra.Command{
	Use:   "required",
	Short: "Check required configuration",
	RunE:  runRequired,
}

func runRequired(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	result, err := service.CheckRequired(database)
	if err != nil {
		outputError(fmt.Errorf("check required: %w", err))
		return nil
	}

	if result.Ready {
		outputJSON(map[string]interface{}{
			"success": true,
			"ready":   true,
			"message": "All required fields configured",
		})
	} else {
		outputJSON(map[string]interface{}{
			"success":          true,
			"ready":            false,
			"missing_required": result.MissingRequired,
			"total_missing":    len(result.MissingRequired),
			"message":          "Please configure required settings",
		})
	}

	return nil
}
