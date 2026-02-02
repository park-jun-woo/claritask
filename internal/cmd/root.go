package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"parkjunwoo.com/talos/internal/db"
)

var rootCmd = &cobra.Command{
	Use:   "talos",
	Short: "Task And LLM Operating System",
	Long:  "Claude Code를 위한 장시간 자동 실행 시스템",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// getDB opens a database connection from the current directory
func getDB() (*db.DB, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(cwd, ".talos", "db")
	return db.Open(dbPath)
}

// outputJSON writes a value as JSON to stdout
func outputJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// outputError writes an error as JSON to stdout
func outputError(err error) {
	outputJSON(map[string]interface{}{
		"success": false,
		"error":   err.Error(),
	})
}

// parseJSON parses a JSON string into a value
func parseJSON(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(techCmd)
	rootCmd.AddCommand(designCmd)
	rootCmd.AddCommand(requiredCmd)
	rootCmd.AddCommand(phaseCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(memoCmd)
}
