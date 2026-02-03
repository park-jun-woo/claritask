package service

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Service provides database access for the bot
type Service struct {
	db *sql.DB
}

// New creates a new service instance
func New(dbPath string) (*Service, error) {
	// Ensure database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database not found: %s", dbPath)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys and WAL mode
	if _, err := db.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, err
	}

	return &Service{db: db}, nil
}

// Close closes the database connection
func (s *Service) Close() error {
	return s.db.Close()
}

// Project represents a project
type Project struct {
	ID          string
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
}

// Task represents a task
type Task struct {
	ID          int64
	FeatureID   int64
	Status      string
	Title       string
	Content     string
	TargetFile  string
	Result      string
	Error       string
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// Feature represents a feature
type Feature struct {
	ID          int64
	ProjectID   string
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
}

// Expert represents an expert
type Expert struct {
	ID          string
	Name        string
	Domain      string
	Status      string
	Description string
}

// Message represents a message
type Message struct {
	ID          int64
	ProjectID   string
	FeatureID   *int64
	Content     string
	Response    string
	Status      string
	Error       string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// ProjectStatus holds project statistics
type ProjectStatus struct {
	TotalTasks      int
	CompletedTasks  int
	InProgressTasks int
	PendingTasks    int
	FailedTasks     int
	Progress        float64
}

// GetCurrentProject returns the current active project
func (s *Service) GetCurrentProject() (*Project, error) {
	// Get current project ID from state
	var projectID string
	err := s.db.QueryRow("SELECT value FROM state WHERE key = 'current_project'").Scan(&projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return s.GetProject(projectID)
}

// GetProject returns a project by ID
func (s *Service) GetProject(id string) (*Project, error) {
	p := &Project{}
	err := s.db.QueryRow(`
		SELECT id, name, description, status, created_at
		FROM projects WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

// ListProjects returns all projects
func (s *Service) ListProjects() ([]Project, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, status, created_at
		FROM projects ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// GetProjectStatus returns task statistics for a project
func (s *Service) GetProjectStatus(projectID string) (*ProjectStatus, error) {
	status := &ProjectStatus{}

	// Get task counts by status
	rows, err := s.db.Query(`
		SELECT t.status, COUNT(*)
		FROM tasks t
		JOIN features f ON t.feature_id = f.id
		WHERE f.project_id = ?
		GROUP BY t.status
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var taskStatus string
		var count int
		if err := rows.Scan(&taskStatus, &count); err != nil {
			return nil, err
		}
		status.TotalTasks += count
		switch taskStatus {
		case "done":
			status.CompletedTasks = count
		case "doing":
			status.InProgressTasks = count
		case "pending":
			status.PendingTasks = count
		case "failed":
			status.FailedTasks = count
		}
	}

	if status.TotalTasks > 0 {
		status.Progress = float64(status.CompletedTasks) / float64(status.TotalTasks) * 100
	}

	return status, rows.Err()
}

// SetCurrentProject sets the current project
func (s *Service) SetCurrentProject(projectID string) error {
	_, err := s.db.Exec(`
		INSERT INTO state (key, value) VALUES ('current_project', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, projectID)
	return err
}

// ListTasks returns tasks for a project with optional status filter
func (s *Service) ListTasks(projectID string, status string, limit int) ([]Task, error) {
	query := `
		SELECT t.id, t.feature_id, t.status, t.title, t.content,
		       COALESCE(t.target_file, ''), COALESCE(t.result, ''), COALESCE(t.error, ''),
		       t.created_at, t.started_at, t.completed_at
		FROM tasks t
		JOIN features f ON t.feature_id = f.id
		WHERE f.project_id = ?
	`
	args := []interface{}{projectID}

	if status != "" {
		query += " AND t.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY t.id DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.FeatureID, &t.Status, &t.Title, &t.Content,
			&t.TargetFile, &t.Result, &t.Error, &t.CreatedAt, &t.StartedAt, &t.CompletedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// GetTask returns a task by ID
func (s *Service) GetTask(id int64) (*Task, error) {
	t := &Task{}
	err := s.db.QueryRow(`
		SELECT id, feature_id, status, title, content,
		       COALESCE(target_file, ''), COALESCE(result, ''), COALESCE(error, ''),
		       created_at, started_at, completed_at
		FROM tasks WHERE id = ?
	`, id).Scan(&t.ID, &t.FeatureID, &t.Status, &t.Title, &t.Content,
		&t.TargetFile, &t.Result, &t.Error, &t.CreatedAt, &t.StartedAt, &t.CompletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

// UpdateTaskStatus updates task status
func (s *Service) UpdateTaskStatus(id int64, status string) error {
	now := time.Now().Format(time.RFC3339)
	var query string

	switch status {
	case "doing":
		query = "UPDATE tasks SET status = ?, started_at = ? WHERE id = ?"
	case "done":
		query = "UPDATE tasks SET status = ?, completed_at = ? WHERE id = ?"
	case "failed":
		query = "UPDATE tasks SET status = ?, failed_at = ? WHERE id = ?"
	default:
		query = "UPDATE tasks SET status = ? WHERE id = ?"
		_, err := s.db.Exec(query, status, id)
		return err
	}

	_, err := s.db.Exec(query, status, now, id)
	return err
}

// ListMessages returns messages for a project
func (s *Service) ListMessages(projectID string, limit int) ([]Message, error) {
	query := `
		SELECT id, project_id, feature_id, content, COALESCE(response, ''),
		       status, COALESCE(error, ''), created_at, completed_at
		FROM messages
		WHERE project_id = ?
		ORDER BY created_at DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.FeatureID, &m.Content, &m.Response,
			&m.Status, &m.Error, &m.CreatedAt, &m.CompletedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

// SendMessage creates a new message
func (s *Service) SendMessage(projectID, content string, featureID *int64) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := s.db.Exec(`
		INSERT INTO messages (project_id, feature_id, content, status, created_at)
		VALUES (?, ?, ?, 'pending', ?)
	`, projectID, featureID, content, now)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ListExperts returns experts assigned to a project
func (s *Service) ListExperts(projectID string) ([]Expert, error) {
	rows, err := s.db.Query(`
		SELECT e.id, e.name, e.domain, e.status, e.description
		FROM experts e
		JOIN project_experts pe ON e.id = pe.expert_id
		WHERE pe.project_id = ?
		ORDER BY e.name
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experts []Expert
	for rows.Next() {
		var e Expert
		if err := rows.Scan(&e.ID, &e.Name, &e.Domain, &e.Status, &e.Description); err != nil {
			return nil, err
		}
		experts = append(experts, e)
	}
	return experts, rows.Err()
}

// GetFeature returns a feature by ID
func (s *Service) GetFeature(id int64) (*Feature, error) {
	f := &Feature{}
	err := s.db.QueryRow(`
		SELECT id, project_id, name, description, status, created_at
		FROM features WHERE id = ?
	`, id).Scan(&f.ID, &f.ProjectID, &f.Name, &f.Description, &f.Status, &f.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

// GetFeatureName returns feature name by ID
func (s *Service) GetFeatureName(id int64) string {
	var name string
	s.db.QueryRow("SELECT name FROM features WHERE id = ?", id).Scan(&name)
	return name
}

// GetExpertDir returns the experts directory path
func (s *Service) GetExpertDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claritask", "experts")
}
