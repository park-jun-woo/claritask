package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"parkjunwoo.com/claritask/internal/service"
)

var edgeCmd = &cobra.Command{
	Use:   "edge",
	Short: "Edge (dependency) management commands",
}

var edgeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a dependency edge",
	RunE:  runEdgeAdd,
}

var edgeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all edges",
	RunE:  runEdgeList,
}

var edgeRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a dependency edge",
	RunE:  runEdgeRemove,
}

var edgeInferCmd = &cobra.Command{
	Use:   "infer",
	Short: "Infer dependency edges using LLM",
	RunE:  runEdgeInfer,
}

func init() {
	edgeCmd.AddCommand(edgeAddCmd)
	edgeCmd.AddCommand(edgeListCmd)
	edgeCmd.AddCommand(edgeRemoveCmd)
	edgeCmd.AddCommand(edgeInferCmd)

	// edge add flags
	edgeAddCmd.Flags().Int64("from", 0, "ID that depends on another (required)")
	edgeAddCmd.Flags().Int64("to", 0, "ID that is depended on (required)")
	edgeAddCmd.Flags().Bool("feature", false, "Create feature edge instead of task edge")
	edgeAddCmd.MarkFlagRequired("from")
	edgeAddCmd.MarkFlagRequired("to")

	// edge list flags
	edgeListCmd.Flags().Bool("feature", false, "Show feature edges only")
	edgeListCmd.Flags().Bool("task", false, "Show task edges only")
	edgeListCmd.Flags().Int64("feature-id", 0, "Filter by feature ID")

	// edge remove flags
	edgeRemoveCmd.Flags().Int64("from", 0, "From ID (required)")
	edgeRemoveCmd.Flags().Int64("to", 0, "To ID (required)")
	edgeRemoveCmd.Flags().Bool("feature", false, "Remove feature edge instead of task edge")
	edgeRemoveCmd.MarkFlagRequired("from")
	edgeRemoveCmd.MarkFlagRequired("to")

	// edge infer flags
	edgeInferCmd.Flags().Int64("feature", 0, "Feature ID to infer task edges within")
	edgeInferCmd.Flags().Bool("project", false, "Infer feature edges at project level")
	edgeInferCmd.Flags().Float64("min-confidence", 0.7, "Minimum confidence threshold for inferred edges")
}

func runEdgeAdd(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	fromID, _ := cmd.Flags().GetInt64("from")
	toID, _ := cmd.Flags().GetInt64("to")
	isFeature, _ := cmd.Flags().GetBool("feature")

	if isFeature {
		// Feature edge
		err := service.AddFeatureEdge(database, fromID, toID)
		if err != nil {
			if err.Error() == "adding edge would create a cycle" {
				hasCycle, path, _ := service.CheckFeatureCycle(database, fromID, toID)
				if hasCycle {
					outputJSON(map[string]interface{}{
						"success": false,
						"error":   "Circular dependency detected",
						"cycle":   path,
					})
					return nil
				}
			}
			outputError(fmt.Errorf("add feature edge: %w", err))
			return nil
		}

		// Get feature names
		fromFeature, _ := service.GetFeature(database, fromID)
		toFeature, _ := service.GetFeature(database, toID)
		fromName := ""
		toName := ""
		if fromFeature != nil {
			fromName = fromFeature.Name
		}
		if toFeature != nil {
			toName = toFeature.Name
		}

		outputJSON(map[string]interface{}{
			"success": true,
			"type":    "feature",
			"from_id": fromID,
			"to_id":   toID,
			"message": fmt.Sprintf("Feature edge created: %s depends on %s", fromName, toName),
		})
	} else {
		// Task edge
		fromIDStr := strconv.FormatInt(fromID, 10)
		toIDStr := strconv.FormatInt(toID, 10)

		err := service.AddTaskEdge(database, fromIDStr, toIDStr)
		if err != nil {
			if err.Error() == "adding edge would create a cycle" {
				hasCycle, path, _ := service.CheckTaskCycle(database, fromIDStr, toIDStr)
				if hasCycle {
					outputJSON(map[string]interface{}{
						"success": false,
						"error":   "Circular dependency detected",
						"cycle":   path,
					})
					return nil
				}
			}
			outputError(fmt.Errorf("add task edge: %w", err))
			return nil
		}

		// Get task titles
		fromTask, _ := service.GetTask(database, fromID)
		toTask, _ := service.GetTask(database, toID)
		fromTitle := ""
		toTitle := ""
		if fromTask != nil {
			fromTitle = fromTask.Title
		}
		if toTask != nil {
			toTitle = toTask.Title
		}

		outputJSON(map[string]interface{}{
			"success": true,
			"type":    "task",
			"from_id": fromID,
			"to_id":   toID,
			"message": fmt.Sprintf("Task edge created: %s depends on %s", fromTitle, toTitle),
		})
	}

	return nil
}

func runEdgeList(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	featureOnly, _ := cmd.Flags().GetBool("feature")
	taskOnly, _ := cmd.Flags().GetBool("task")
	filterFeatureID, _ := cmd.Flags().GetInt64("feature-id")

	result, err := service.ListAllEdges(database)
	if err != nil {
		outputError(fmt.Errorf("list edges: %w", err))
		return nil
	}

	response := map[string]interface{}{
		"success": true,
	}

	if !taskOnly {
		response["feature_edges"] = result.FeatureEdges
		response["total_feature_edges"] = result.TotalFeature
	}

	if !featureOnly {
		if filterFeatureID > 0 {
			// Filter task edges by feature
			edges, err := service.GetTaskEdgesByFeature(database, filterFeatureID)
			if err != nil {
				outputError(fmt.Errorf("get task edges by feature: %w", err))
				return nil
			}

			var taskEdges []service.TaskEdgeItem
			for _, e := range edges {
				fromTask, _ := service.GetTask(database, e.FromTaskID)
				toTask, _ := service.GetTask(database, e.ToTaskID)
				fromTitle := ""
				toTitle := ""
				if fromTask != nil {
					fromTitle = fromTask.Title
				}
				if toTask != nil {
					toTitle = toTask.Title
				}
				taskEdges = append(taskEdges, service.TaskEdgeItem{
					From: service.TaskRef{ID: e.FromTaskID, Title: fromTitle},
					To:   service.TaskRef{ID: e.ToTaskID, Title: toTitle},
				})
			}
			response["task_edges"] = taskEdges
			response["total_task_edges"] = len(taskEdges)
		} else {
			response["task_edges"] = result.TaskEdges
			response["total_task_edges"] = result.TotalTask
		}
	}

	outputJSON(response)
	return nil
}

func runEdgeRemove(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	fromID, _ := cmd.Flags().GetInt64("from")
	toID, _ := cmd.Flags().GetInt64("to")
	isFeature, _ := cmd.Flags().GetBool("feature")

	if isFeature {
		err := service.RemoveFeatureEdge(database, fromID, toID)
		if err != nil {
			outputError(fmt.Errorf("remove feature edge: %w", err))
			return nil
		}

		outputJSON(map[string]interface{}{
			"success": true,
			"type":    "feature",
			"from_id": fromID,
			"to_id":   toID,
			"message": "Edge removed successfully",
		})
	} else {
		fromIDStr := strconv.FormatInt(fromID, 10)
		toIDStr := strconv.FormatInt(toID, 10)

		err := service.RemoveTaskEdge(database, fromIDStr, toIDStr)
		if err != nil {
			outputError(fmt.Errorf("remove task edge: %w", err))
			return nil
		}

		outputJSON(map[string]interface{}{
			"success": true,
			"type":    "task",
			"from_id": fromID,
			"to_id":   toID,
			"message": "Edge removed successfully",
		})
	}

	return nil
}

func runEdgeInfer(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	featureID, _ := cmd.Flags().GetInt64("feature")
	projectLevel, _ := cmd.Flags().GetBool("project")
	minConfidence, _ := cmd.Flags().GetFloat64("min-confidence")

	if featureID == 0 && !projectLevel {
		outputError(fmt.Errorf("specify --feature <id> or --project"))
		return nil
	}

	if featureID > 0 && projectLevel {
		outputError(fmt.Errorf("cannot specify both --feature and --project"))
		return nil
	}

	if featureID > 0 {
		// Infer task edges within a feature
		ctx, err := service.PrepareTaskEdgeInference(database, featureID)
		if err != nil {
			outputError(fmt.Errorf("prepare task edge inference: %w", err))
			return nil
		}

		outputJSON(map[string]interface{}{
			"success":        true,
			"type":           "task",
			"feature_id":     featureID,
			"items":          ctx.Items,
			"existing_edges": ctx.Existing,
			"prompt":         ctx.Prompt,
			"min_confidence": minConfidence,
			"instructions":   "Use the prompt to infer edges, then run 'clari edge add --from <id> --to <id>' for each inferred edge",
		})
	} else {
		// Infer feature edges at project level
		ctx, err := service.PrepareFeatureEdgeInference(database)
		if err != nil {
			outputError(fmt.Errorf("prepare feature edge inference: %w", err))
			return nil
		}

		outputJSON(map[string]interface{}{
			"success":        true,
			"type":           "feature",
			"items":          ctx.Items,
			"existing_edges": ctx.Existing,
			"prompt":         ctx.Prompt,
			"min_confidence": minConfidence,
			"instructions":   "Use the prompt to infer edges, then run 'clari edge add --feature --from <id> --to <id>' for each inferred edge",
		})
	}

	return nil
}
