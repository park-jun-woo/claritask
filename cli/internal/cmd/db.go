package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"parkjunwoo.com/claritask/internal/db"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
}

var dbVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current database version",
	RunE:  runDBVersion,
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE:  runDBMigrate,
}

var dbRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback to specific version",
	RunE:  runDBRollback,
}

var dbBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create database backup",
	RunE:  runDBBackup,
}

var rollbackVersion int

func init() {
	dbCmd.AddCommand(dbVersionCmd)
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbRollbackCmd)
	dbCmd.AddCommand(dbBackupCmd)

	dbRollbackCmd.Flags().IntVar(&rollbackVersion, "version", 0, "Target version to rollback")
	dbRollbackCmd.MarkFlagRequired("version")
}

func runDBVersion(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	version := database.GetVersion()
	outputJSON(map[string]interface{}{
		"success":         true,
		"current_version": version,
		"latest_version":  db.LatestVersion,
		"up_to_date":      version >= db.LatestVersion,
	})
	return nil
}

func runDBMigrate(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	beforeVersion := database.GetVersion()
	if err := database.AutoMigrate(); err != nil {
		outputError(fmt.Errorf("migrate: %w", err))
		return nil
	}
	afterVersion := database.GetVersion()

	outputJSON(map[string]interface{}{
		"success":            true,
		"before_version":     beforeVersion,
		"after_version":      afterVersion,
		"migrations_applied": afterVersion - beforeVersion,
		"message":            "Migration completed",
	})
	return nil
}

func runDBRollback(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	beforeVersion := database.GetVersion()
	if err := database.Rollback(rollbackVersion); err != nil {
		outputError(fmt.Errorf("rollback: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":        true,
		"before_version": beforeVersion,
		"after_version":  rollbackVersion,
		"message":        "Rollback completed",
	})
	return nil
}

func runDBBackup(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		outputError(fmt.Errorf("open database: %w", err))
		return nil
	}
	defer database.Close()

	backupPath, err := database.Backup()
	if err != nil {
		outputError(fmt.Errorf("backup: %w", err))
		return nil
	}

	outputJSON(map[string]interface{}{
		"success":     true,
		"backup_path": backupPath,
		"message":     "Backup created successfully",
	})
	return nil
}
