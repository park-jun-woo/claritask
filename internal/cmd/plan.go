package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"parkjunwoo.com/claritask/internal/service"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Planning commands for project analysis",
}

var planFeaturesCmd = &cobra.Command{
	Use:   "features",
	Short: "Generate feature list from project description",
	RunE:  runPlanFeatures,
}

func init() {
	planCmd.AddCommand(planFeaturesCmd)

	// plan features flags
	planFeaturesCmd.Flags().Bool("auto-create", false, "Automatically create suggested features")
}

func runPlanFeatures(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	autoCreate, _ := cmd.Flags().GetBool("auto-create")

	// Prepare feature planning context
	plan, err := service.PreparePlanFeatures(database)
	if err != nil {
		outputError(fmt.Errorf("prepare feature plan: %w", err))
		return nil
	}

	response := map[string]interface{}{
		"success":      true,
		"prompt":       plan.Prompt,
		"instructions": "Use the prompt to generate features, then run 'clari feature add' for each planned feature",
	}

	// Check for JSON input with planned features
	if len(args) > 0 {
		var plannedFeatures struct {
			Features  []service.PlannedFeature `json:"features"`
			Reasoning string                   `json:"reasoning"`
		}

		if err := json.Unmarshal([]byte(args[0]), &plannedFeatures); err == nil && len(plannedFeatures.Features) > 0 {
			// Features were provided
			plan.Features = plannedFeatures.Features
			plan.TotalCount = len(plannedFeatures.Features)
			plan.Reasoning = plannedFeatures.Reasoning

			response["features"] = plan.Features
			response["total_count"] = plan.TotalCount
			response["reasoning"] = plan.Reasoning

			if autoCreate {
				// Get project
				project, err := service.GetProject(database)
				if err != nil {
					outputError(fmt.Errorf("get project: %w", err))
					return nil
				}

				var created []map[string]interface{}
				for _, f := range plan.Features {
					featureID, err := service.CreateFeature(database, project.ID, f.Name, f.Description)
					if err != nil {
						continue
					}

					// Set spec if description provided
					if f.Description != "" {
						service.SetFeatureSpec(database, featureID, f.Description)
					}

					created = append(created, map[string]interface{}{
						"id":          featureID,
						"name":        f.Name,
						"description": f.Description,
						"priority":    f.Priority,
					})
				}

				response["created_features"] = created
				response["total_created"] = len(created)
				response["message"] = fmt.Sprintf("Created %d features automatically", len(created))
			}

			delete(response, "prompt")
			delete(response, "instructions")
		}
	}

	outputJSON(response)
	return nil
}
