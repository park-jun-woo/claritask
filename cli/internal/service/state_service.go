package service

import (
	"database/sql"
	"fmt"
	"strconv"

	"parkjunwoo.com/claritask/internal/db"
)

// State key constants
const (
	StateCurrentProject = "current_project"
	StateCurrentFeature = "current_feature"
	StateCurrentTask    = "current_task"
	StateNextTask       = "next_task"
)

// SetState sets a state value (upsert)
func SetState(database *db.DB, key, value string) error {
	_, err := database.Exec(
		`INSERT INTO state (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = ?`,
		key, value, value,
	)
	if err != nil {
		return fmt.Errorf("set state: %w", err)
	}
	return nil
}

// GetState retrieves a state value
func GetState(database *db.DB, key string) (string, error) {
	row := database.QueryRow(`SELECT value FROM state WHERE key = ?`, key)
	var value string
	err := row.Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("get state: %w", err)
	}
	return value, nil
}

// GetAllStates retrieves all state values
func GetAllStates(database *db.DB) (map[string]string, error) {
	rows, err := database.Query(`SELECT key, value FROM state`)
	if err != nil {
		return nil, fmt.Errorf("get all states: %w", err)
	}
	defer rows.Close()

	states := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan state: %w", err)
		}
		states[key] = value
	}
	return states, nil
}

// DeleteState deletes a state value
func DeleteState(database *db.DB, key string) error {
	result, err := database.Exec(`DELETE FROM state WHERE key = ?`, key)
	if err != nil {
		return fmt.Errorf("delete state: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("state not found: %s", key)
	}
	return nil
}

// UpdateCurrentState updates the current state when a task is executed
func UpdateCurrentState(database *db.DB, projectID string, featureID, taskID, nextTaskID int64) error {
	if projectID != "" {
		if err := SetState(database, StateCurrentProject, projectID); err != nil {
			return err
		}
	}
	if featureID > 0 {
		if err := SetState(database, StateCurrentFeature, strconv.FormatInt(featureID, 10)); err != nil {
			return err
		}
	}
	if taskID > 0 {
		if err := SetState(database, StateCurrentTask, strconv.FormatInt(taskID, 10)); err != nil {
			return err
		}
	}
	if nextTaskID > 0 {
		if err := SetState(database, StateNextTask, strconv.FormatInt(nextTaskID, 10)); err != nil {
			return err
		}
	} else {
		// No next task
		SetState(database, StateNextTask, "")
	}
	return nil
}

// InitializeProjectState initializes state for a new project
func InitializeProjectState(database *db.DB, projectID string) error {
	if err := SetState(database, StateCurrentProject, projectID); err != nil {
		return err
	}
	if err := SetState(database, StateCurrentFeature, ""); err != nil {
		return err
	}
	if err := SetState(database, StateCurrentTask, ""); err != nil {
		return err
	}
	if err := SetState(database, StateNextTask, ""); err != nil {
		return err
	}
	return nil
}
