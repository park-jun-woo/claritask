package service

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// CreateMessage creates a new message
func CreateMessage(database *db.DB, projectID string, featureID *int64, content string) (*model.Message, error) {
	now := db.TimeNow()

	var result sql.Result
	var err error

	if featureID != nil && *featureID > 0 {
		result, err = database.Exec(
			`INSERT INTO messages (project_id, feature_id, content, status, created_at)
			 VALUES (?, ?, ?, 'pending', ?)`,
			projectID, *featureID, content, now,
		)
	} else {
		result, err = database.Exec(
			`INSERT INTO messages (project_id, content, status, created_at)
			 VALUES (?, ?, 'pending', ?)`,
			projectID, content, now,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get message id: %w", err)
	}

	createdAt, _ := db.ParseTime(now)
	return &model.Message{
		ID:        id,
		ProjectID: projectID,
		FeatureID: featureID,
		Content:   content,
		Status:    "pending",
		CreatedAt: createdAt,
	}, nil
}

// GetMessage retrieves a message by ID with associated tasks
func GetMessage(database *db.DB, messageID int64) (*model.MessageDetail, error) {
	row := database.QueryRow(
		`SELECT id, project_id, feature_id, content, response, status, error, created_at, completed_at
		 FROM messages WHERE id = ?`,
		messageID,
	)

	var m model.MessageDetail
	var featureID sql.NullInt64
	var response, errMsg sql.NullString
	var createdAt string
	var completedAt sql.NullString

	err := row.Scan(&m.ID, &m.Content, &featureID, &m.Content, &response, &m.Status, &errMsg, &createdAt, &completedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("get message: %w", err)
	}

	if featureID.Valid {
		m.FeatureID = &featureID.Int64
	}
	if response.Valid {
		m.Response = response.String
	}
	if errMsg.Valid {
		m.Error = errMsg.String
	}
	m.CreatedAt = createdAt
	if completedAt.Valid {
		m.CompletedAt = &completedAt.String
	}

	// Get associated tasks
	rows, err := database.Query(
		`SELECT t.id, t.title, t.status
		 FROM tasks t
		 INNER JOIN message_tasks mt ON t.id = mt.task_id
		 WHERE mt.message_id = ?
		 ORDER BY t.id`,
		messageID,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("get message tasks: %w", err)
	}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var task model.TaskListItem
			if err := rows.Scan(&task.ID, &task.Title, &task.Status); err != nil {
				return nil, fmt.Errorf("scan task: %w", err)
			}
			m.Tasks = append(m.Tasks, task)
		}
	}

	return &m, nil
}

// ListMessages lists messages with optional filters
func ListMessages(database *db.DB, projectID string, status string, featureID *int64, limit int) ([]model.MessageListItem, int, error) {
	query := `SELECT m.id, m.content, m.status, m.feature_id, m.created_at,
	          (SELECT COUNT(*) FROM message_tasks mt WHERE mt.message_id = m.id) as tasks_count
	          FROM messages m WHERE m.project_id = ?`
	args := []interface{}{projectID}

	if status != "" {
		query += " AND m.status = ?"
		args = append(args, status)
	}

	if featureID != nil && *featureID > 0 {
		query += " AND m.feature_id = ?"
		args = append(args, *featureID)
	}

	query += " ORDER BY m.created_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []model.MessageListItem
	for rows.Next() {
		var m model.MessageListItem
		var featureID sql.NullInt64
		if err := rows.Scan(&m.ID, &m.Content, &m.Status, &featureID, &m.CreatedAt, &m.TasksCount); err != nil {
			return nil, 0, fmt.Errorf("scan message: %w", err)
		}
		if featureID.Valid {
			m.FeatureID = &featureID.Int64
		}
		messages = append(messages, m)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM messages WHERE project_id = ?`
	countArgs := []interface{}{projectID}
	if status != "" {
		countQuery += " AND status = ?"
		countArgs = append(countArgs, status)
	}
	if featureID != nil && *featureID > 0 {
		countQuery += " AND feature_id = ?"
		countArgs = append(countArgs, *featureID)
	}

	var total int
	database.QueryRow(countQuery, countArgs...).Scan(&total)

	return messages, total, nil
}

// UpdateMessageStatus updates a message's status
func UpdateMessageStatus(database *db.DB, messageID int64, status, response, errMsg string) error {
	var completedAt interface{}
	if status == "completed" || status == "failed" {
		completedAt = db.TimeNow()
	}

	_, err := database.Exec(
		`UPDATE messages SET status = ?, response = ?, error = ?, completed_at = ? WHERE id = ?`,
		status, response, errMsg, completedAt, messageID,
	)
	if err != nil {
		return fmt.Errorf("update message status: %w", err)
	}
	return nil
}

// DeleteMessage deletes a message
func DeleteMessage(database *db.DB, messageID int64) error {
	result, err := database.Exec(`DELETE FROM messages WHERE id = ?`, messageID)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// LinkMessageTask links a message to a task
func LinkMessageTask(database *db.DB, messageID, taskID int64) error {
	now := db.TimeNow()
	_, err := database.Exec(
		`INSERT INTO message_tasks (message_id, task_id, created_at) VALUES (?, ?, ?)`,
		messageID, taskID, now,
	)
	if err != nil {
		return fmt.Errorf("link message task: %w", err)
	}
	return nil
}

// CountMessageTasks counts tasks linked to a message
func CountMessageTasks(database *db.DB, messageID int64) (int, error) {
	var count int
	err := database.QueryRow(
		`SELECT COUNT(*) FROM message_tasks WHERE message_id = ?`,
		messageID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count message tasks: %w", err)
	}
	return count, nil
}

// MessageAnalysisSystemPrompt returns system prompt for Claude Code message analysis mode
func MessageAnalysisSystemPrompt() string {
	return `You are in Claritask Message Analysis Mode.

ROLE: Analyze user's modification request and create Tasks.

WORKFLOW:
1. Analyze the user's request
2. Break down into actionable Tasks
3. Register each Task using: clari task add '{"feature_id": N, "title": "...", "content": "..."}'
4. Write a summary report

COMPLETION:
When all Tasks are registered, create '.claritask/complete' file with the report content.
Report format (Markdown):
- Summary of request analysis
- List of created Tasks with IDs
- Any assumptions or clarifications needed

CONSTRAINTS:
- Always register Tasks, never implement directly
- Each Task should be specific and actionable
- Reference existing FDL if available
- Link Tasks to appropriate Feature`
}

// RunMessageAnalysisWithTTY runs message analysis with TTY handover
func RunMessageAnalysisWithTTY(database *db.DB, message *model.Message) (string, int, error) {
	// Update status to processing
	if err := UpdateMessageStatus(database, message.ID, "processing", "", ""); err != nil {
		return "", 0, fmt.Errorf("update status: %w", err)
	}

	// Ensure .claritask directory exists
	claritaskDir := ".claritask"
	if err := os.MkdirAll(claritaskDir, 0755); err != nil {
		return "", 0, fmt.Errorf("create .claritask directory: %w", err)
	}

	// Remove any existing complete file
	completeFile := filepath.Join(claritaskDir, "complete")
	os.Remove(completeFile)

	// Build prompts
	systemPrompt := MessageAnalysisSystemPrompt()
	initialPrompt := BuildMessageAnalysisPrompt(database, message)

	// Run TTY handover
	err := RunWithTTYHandoverEx(systemPrompt, initialPrompt, "acceptEdits", completeFile)

	// Read complete file content for response
	var response string
	if content, readErr := os.ReadFile(completeFile); readErr == nil {
		response = string(content)
	}

	// Cleanup complete file
	os.Remove(completeFile)

	if err != nil {
		UpdateMessageStatus(database, message.ID, "failed", "", err.Error())
		return "", 0, fmt.Errorf("TTY handover failed: %w", err)
	}

	// Count created tasks
	tasksCount, _ := CountMessageTasks(database, message.ID)

	// Update status to completed
	UpdateMessageStatus(database, message.ID, "completed", response, "")

	// Save report
	reportPath, _ := SaveMessageReport(message.ID, response)

	return reportPath, tasksCount, nil
}

// BuildMessageAnalysisPrompt builds the initial prompt for message analysis
func BuildMessageAnalysisPrompt(database *db.DB, message *model.Message) string {
	var featureInfo string
	if message.FeatureID != nil && *message.FeatureID > 0 {
		if feature, err := GetFeature(database, *message.FeatureID); err == nil && feature != nil {
			featureInfo = fmt.Sprintf(`
=== Related Feature ===
ID: %d
Name: %s
Description: %s
FDL Available: %v
`, feature.ID, feature.Name, feature.Description, feature.FDL != "")
		}
	}

	return fmt.Sprintf(`[CLARITASK MESSAGE ANALYSIS]

Message ID: %d
Project: %s

=== User Request ===
%s
%s
---

Analyze the request and create Tasks.

IMPORTANT:
- Use "clari task add" to register each task
- After creating all tasks, create '.claritask/complete' with a report
`, message.ID, message.ProjectID, message.Content, featureInfo)
}

// SaveMessageReport saves the analysis report to reports folder
func SaveMessageReport(messageID int64, content string) (string, error) {
	if content == "" {
		return "", nil
	}

	// Ensure reports directory exists
	reportsDir := "reports"
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", fmt.Errorf("create reports directory: %w", err)
	}

	// Generate report filename
	timestamp := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-message-%03d.md", timestamp, messageID)
	reportPath := filepath.Join(reportsDir, filename)

	// Write report
	if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write report: %w", err)
	}

	return reportPath, nil
}
