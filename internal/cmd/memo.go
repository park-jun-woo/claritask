package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"parkjunwoo.com/talos/internal/service"
)

var memoCmd = &cobra.Command{
	Use:   "memo",
	Short: "Memo management commands",
}

var memoSetCmd = &cobra.Command{
	Use:   "set <key> '<json>'",
	Short: "Set a memo",
	Args:  cobra.ExactArgs(2),
	RunE:  runMemoSet,
}

var memoGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a memo",
	Args:  cobra.ExactArgs(1),
	RunE:  runMemoGet,
}

var memoListCmd = &cobra.Command{
	Use:   "list [scope]",
	Short: "List memos",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runMemoList,
}

var memoDelCmd = &cobra.Command{
	Use:   "del <key>",
	Short: "Delete a memo",
	Args:  cobra.ExactArgs(1),
	RunE:  runMemoDel,
}

func init() {
	memoCmd.AddCommand(memoSetCmd)
	memoCmd.AddCommand(memoGetCmd)
	memoCmd.AddCommand(memoListCmd)
	memoCmd.AddCommand(memoDelCmd)
}

type memoSetInput struct {
	Value    string   `json:"value"`
	Priority int      `json:"priority"`
	Summary  string   `json:"summary"`
	Tags     []string `json:"tags"`
}

func runMemoSet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	scope, scopeID, key, err := parseKey(args[0])
	if err != nil {
		outputError(err)
		return nil
	}

	var input memoSetInput
	if err := parseJSON(args[1], &input); err != nil {
		outputError(fmt.Errorf("parse JSON: %w", err))
		return nil
	}

	if input.Value == "" {
		outputError(fmt.Errorf("missing required field: value"))
		return nil
	}

	memoInput := service.MemoSetInput{
		Scope:    scope,
		ScopeID:  scopeID,
		Key:      key,
		Value:    input.Value,
		Priority: input.Priority,
		Summary:  input.Summary,
		Tags:     input.Tags,
	}

	if err := service.SetMemo(database, memoInput); err != nil {
		outputError(fmt.Errorf("set memo: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":  true,
		"scope":    scope,
		"scope_id": scopeID,
		"key":      key,
		"message":  "Memo saved successfully",
	})

	return nil
}

func runMemoGet(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	scope, scopeID, key, err := parseKey(args[0])
	if err != nil {
		outputError(err)
		return nil
	}

	memo, err := service.GetMemo(database, scope, scopeID, key)
	if err != nil {
		outputError(fmt.Errorf("get memo: %w", err))
		return nil
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(memo.Data), &data)

	outputJSON(map[string]interface{}{
		"success":    true,
		"scope":      memo.Scope,
		"scope_id":   memo.ScopeID,
		"key":        memo.Key,
		"data":       data,
		"priority":   memo.Priority,
		"created_at": memo.CreatedAt,
		"updated_at": memo.UpdatedAt,
	})

	return nil
}

func runMemoList(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	if len(args) == 0 {
		// List all memos
		result, err := service.ListMemos(database)
		if err != nil {
			outputError(fmt.Errorf("list memos: %w", err))
			return nil
		}

		outputJSON(map[string]interface{}{
			"success": true,
			"memos":   result,
		})
		return nil
	}

	// Parse scope
	scope, scopeID := parseScopeFilter(args[0])

	memos, err := service.ListMemosByScope(database, scope, scopeID)
	if err != nil {
		outputError(fmt.Errorf("list memos by scope: %w", err))
		return nil
	}

	var memoList []map[string]interface{}
	for _, m := range memos {
		var data map[string]interface{}
		json.Unmarshal([]byte(m.Data), &data)
		memoList = append(memoList, map[string]interface{}{
			"key":      m.Key,
			"data":     data,
			"priority": m.Priority,
		})
	}

	outputJSON(map[string]interface{}{
		"success":  true,
		"scope":    scope,
		"scope_id": scopeID,
		"memos":    memoList,
		"total":    len(memoList),
	})

	return nil
}

func runMemoDel(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	scope, scopeID, key, err := parseKey(args[0])
	if err != nil {
		outputError(err)
		return nil
	}

	if err := service.DeleteMemo(database, scope, scopeID, key); err != nil {
		outputError(fmt.Errorf("delete memo: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success": true,
		"message": "Memo deleted successfully",
	})

	return nil
}

// parseKey parses a memo key in the format "1:42:notes" -> (scope, scopeID, key)
func parseKey(input string) (scope, scopeID, key string, err error) {
	parts := strings.Split(input, ":")
	switch len(parts) {
	case 1:
		return "project", "", parts[0], nil
	case 2:
		return "phase", parts[0], parts[1], nil
	case 3:
		return "task", parts[0] + ":" + parts[1], parts[2], nil
	default:
		return "", "", "", fmt.Errorf("invalid key format: %s", input)
	}
}

// parseScopeFilter parses a scope filter like "1" or "1:42"
func parseScopeFilter(input string) (scope, scopeID string) {
	parts := strings.Split(input, ":")
	switch len(parts) {
	case 1:
		return "phase", parts[0]
	case 2:
		return "task", parts[0] + ":" + parts[1]
	default:
		return "project", ""
	}
}
