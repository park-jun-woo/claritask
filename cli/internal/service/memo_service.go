package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// MemoSetInput represents input for setting a memo
type MemoSetInput struct {
	Scope    string   // "project", "feature", "task"
	ScopeID  string   // project_id, feature_id, task_id
	Key      string
	Value    string
	Priority int // 1, 2, 3
	Summary  string
	Tags     []string
}

// SetMemo sets a memo (upsert)
func SetMemo(database *db.DB, input MemoSetInput) error {
	data := map[string]interface{}{
		"value": input.Value,
	}
	if input.Summary != "" {
		data["summary"] = input.Summary
	}
	if len(input.Tags) > 0 {
		data["tags"] = input.Tags
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal memo data: %w", err)
	}

	priority := input.Priority
	if priority == 0 {
		priority = 2 // default
	}

	now := db.TimeNow()
	_, err = database.Exec(
		`INSERT INTO memos (scope, scope_id, key, data, priority, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(scope, scope_id, key) DO UPDATE SET data = ?, priority = ?, updated_at = ?`,
		input.Scope, input.ScopeID, input.Key, string(jsonData), priority, now, now,
		string(jsonData), priority, now,
	)
	if err != nil {
		return fmt.Errorf("set memo: %w", err)
	}
	return nil
}

// GetMemo retrieves a memo by scope, scope_id, and key
func GetMemo(database *db.DB, scope, scopeID, key string) (*model.Memo, error) {
	row := database.QueryRow(
		`SELECT scope, scope_id, key, data, priority, created_at, updated_at
		 FROM memos WHERE scope = ? AND scope_id = ? AND key = ?`,
		scope, scopeID, key,
	)
	var m model.Memo
	var createdAt, updatedAt string
	err := row.Scan(&m.Scope, &m.ScopeID, &m.Key, &m.Data, &m.Priority, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("get memo: %w", err)
	}
	m.CreatedAt, _ = db.ParseTime(createdAt)
	m.UpdatedAt, _ = db.ParseTime(updatedAt)
	return &m, nil
}

// DeleteMemo deletes a memo
func DeleteMemo(database *db.DB, scope, scopeID, key string) error {
	result, err := database.Exec(
		`DELETE FROM memos WHERE scope = ? AND scope_id = ? AND key = ?`,
		scope, scopeID, key,
	)
	if err != nil {
		return fmt.Errorf("delete memo: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("memo not found")
	}
	return nil
}

// MemoSummary represents a memo summary
type MemoSummary struct {
	Key      string `json:"key"`
	Priority int    `json:"priority"`
	Summary  string `json:"summary,omitempty"`
}

// MemoListResult represents the result of ListMemos
type MemoListResult struct {
	Project map[string][]MemoSummary `json:"project"`
	Feature map[string][]MemoSummary `json:"feature"`
	Task    map[string][]MemoSummary `json:"task"`
	Total   int                      `json:"total"`
}

// ListMemos lists all memos
func ListMemos(database *db.DB) (*MemoListResult, error) {
	rows, err := database.Query(
		`SELECT scope, scope_id, key, data, priority FROM memos ORDER BY scope, scope_id, key`,
	)
	if err != nil {
		return nil, fmt.Errorf("list memos: %w", err)
	}
	defer rows.Close()

	result := &MemoListResult{
		Project: make(map[string][]MemoSummary),
		Feature: make(map[string][]MemoSummary),
		Task:    make(map[string][]MemoSummary),
	}

	for rows.Next() {
		var scope, scopeID, key, data string
		var priority int
		if err := rows.Scan(&scope, &scopeID, &key, &data, &priority); err != nil {
			return nil, fmt.Errorf("scan memo: %w", err)
		}

		var dataMap map[string]interface{}
		json.Unmarshal([]byte(data), &dataMap)

		summary := ""
		if s, ok := dataMap["summary"].(string); ok {
			summary = s
		}

		ms := MemoSummary{
			Key:      key,
			Priority: priority,
			Summary:  summary,
		}

		switch scope {
		case "project":
			result.Project[scopeID] = append(result.Project[scopeID], ms)
		case "feature":
			result.Feature[scopeID] = append(result.Feature[scopeID], ms)
		case "task":
			result.Task[scopeID] = append(result.Task[scopeID], ms)
		}
		result.Total++
	}

	return result, nil
}

// ListMemosByScope lists memos by scope
func ListMemosByScope(database *db.DB, scope, scopeID string) ([]model.Memo, error) {
	rows, err := database.Query(
		`SELECT scope, scope_id, key, data, priority, created_at, updated_at
		 FROM memos WHERE scope = ? AND scope_id = ? ORDER BY key`,
		scope, scopeID,
	)
	if err != nil {
		return nil, fmt.Errorf("list memos by scope: %w", err)
	}
	defer rows.Close()

	var memos []model.Memo
	for rows.Next() {
		var m model.Memo
		var createdAt, updatedAt string
		if err := rows.Scan(&m.Scope, &m.ScopeID, &m.Key, &m.Data, &m.Priority, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan memo: %w", err)
		}
		m.CreatedAt, _ = db.ParseTime(createdAt)
		m.UpdatedAt, _ = db.ParseTime(updatedAt)
		memos = append(memos, m)
	}
	return memos, nil
}

// GetHighPriorityMemos returns memos with priority=1
func GetHighPriorityMemos(database *db.DB) ([]model.Memo, error) {
	rows, err := database.Query(
		`SELECT scope, scope_id, key, data, priority, created_at, updated_at
		 FROM memos WHERE priority = 1 ORDER BY scope, scope_id, key`,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return []model.Memo{}, nil
		}
		return nil, fmt.Errorf("get high priority memos: %w", err)
	}
	defer rows.Close()

	var memos []model.Memo
	for rows.Next() {
		var m model.Memo
		var createdAt, updatedAt string
		if err := rows.Scan(&m.Scope, &m.ScopeID, &m.Key, &m.Data, &m.Priority, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan memo: %w", err)
		}
		m.CreatedAt, _ = db.ParseTime(createdAt)
		m.UpdatedAt, _ = db.ParseTime(updatedAt)
		memos = append(memos, m)
	}
	return memos, nil
}

// ParseMemoKey parses a memo key in the format "feature_id:task_id:key"
func ParseMemoKey(input string) (scope, scopeID, key string, err error) {
	parts := strings.Split(input, ":")

	switch len(parts) {
	case 1:
		// project level: "key"
		return "project", "", parts[0], nil
	case 2:
		// feature level: "1:key"
		return "feature", parts[0], parts[1], nil
	case 3:
		// task level: "1:2:key" (feature_id:task_id:key)
		return "task", parts[0] + ":" + parts[1], parts[2], nil
	default:
		return "", "", "", fmt.Errorf("invalid memo key format: %s", input)
	}
}
