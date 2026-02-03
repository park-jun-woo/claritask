package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"parkjunwoo.com/claritask/internal/service"
)

var messageCmd = &cobra.Command{
	Use:   "message",
	Short: "Message management commands",
}

var messageSendCmd = &cobra.Command{
	Use:   "send <content>",
	Short: "Send a modification request and convert to tasks",
	Args:  cobra.ExactArgs(1),
	RunE:  runMessageSend,
}

var messageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List messages",
	RunE:  runMessageList,
}

var messageGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get message details",
	Args:  cobra.ExactArgs(1),
	RunE:  runMessageGet,
}

var messageDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a message",
	Args:  cobra.ExactArgs(1),
	RunE:  runMessageDelete,
}

func init() {
	messageCmd.AddCommand(messageSendCmd)
	messageCmd.AddCommand(messageListCmd)
	messageCmd.AddCommand(messageGetCmd)
	messageCmd.AddCommand(messageDeleteCmd)

	// send flags
	messageSendCmd.Flags().Int64P("feature", "f", 0, "Related feature ID")

	// list flags
	messageListCmd.Flags().StringP("status", "s", "", "Filter by status (pending, processing, completed, failed)")
	messageListCmd.Flags().Int64P("feature", "f", 0, "Filter by feature ID")
	messageListCmd.Flags().IntP("limit", "l", 20, "Max number of results")
}

func runMessageSend(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Get current project
	projectID, err := service.GetState(database, service.StateCurrentProject)
	if err != nil {
		outputError(fmt.Errorf("get current project: %w", err))
		return nil
	}
	if projectID == "" {
		outputError(fmt.Errorf("no project selected. use 'clari project set <id>' first"))
		return nil
	}

	content := args[0]
	featureID, _ := cmd.Flags().GetInt64("feature")

	var featureIDPtr *int64
	if featureID > 0 {
		featureIDPtr = &featureID
	}

	// Create message
	message, err := service.CreateMessage(database, projectID, featureIDPtr, content)
	if err != nil {
		outputError(fmt.Errorf("create message: %w", err))
		return nil
	}

	// Run TTY handover for Claude analysis
	reportPath, tasksCreated, err := service.RunMessageAnalysisWithTTY(database, message)
	if err != nil {
		outputError(fmt.Errorf("message analysis failed: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":       true,
		"message_id":    message.ID,
		"tasks_created": tasksCreated,
		"report_path":   reportPath,
	})

	return nil
}

func runMessageList(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	// Get current project
	projectID, err := service.GetState(database, service.StateCurrentProject)
	if err != nil {
		outputError(fmt.Errorf("get current project: %w", err))
		return nil
	}
	if projectID == "" {
		outputError(fmt.Errorf("no project selected. use 'clari project set <id>' first"))
		return nil
	}

	status, _ := cmd.Flags().GetString("status")
	featureID, _ := cmd.Flags().GetInt64("feature")
	limit, _ := cmd.Flags().GetInt("limit")

	var featureIDPtr *int64
	if featureID > 0 {
		featureIDPtr = &featureID
	}

	messages, total, err := service.ListMessages(database, projectID, status, featureIDPtr, limit)
	if err != nil {
		outputError(fmt.Errorf("list messages: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":  true,
		"messages": messages,
		"total":    total,
	})

	return nil
}

func runMessageGet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	messageID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid message ID: %w", err))
		return nil
	}

	message, err := service.GetMessage(database, messageID)
	if err != nil {
		outputError(fmt.Errorf("get message: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"message": message,
	})

	return nil
}

func runMessageDelete(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	messageID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		outputError(fmt.Errorf("invalid message ID: %w", err))
		return nil
	}

	if err := service.DeleteMessage(database, messageID); err != nil {
		outputError(fmt.Errorf("delete message: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":    true,
		"deleted_id": messageID,
	})

	return nil
}
