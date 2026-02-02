package service

import (
	"database/sql"
	"fmt"
	"strconv"

	"parkjunwoo.com/talos/internal/db"
	"parkjunwoo.com/talos/internal/model"
)

// PhaseCreateInput represents input for creating a phase
type PhaseCreateInput struct {
	ProjectID   string
	Name        string
	Description string
	OrderNum    int
}

// CreatePhase creates a new phase
func CreatePhase(database *db.DB, input PhaseCreateInput) (int64, error) {
	now := db.TimeNow()
	result, err := database.Exec(
		`INSERT INTO phases (project_id, name, description, order_num, status, created_at)
		 VALUES (?, ?, ?, ?, 'pending', ?)`,
		input.ProjectID, input.Name, input.Description, input.OrderNum, now,
	)
	if err != nil {
		return 0, fmt.Errorf("create phase: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}
	return id, nil
}

// GetPhase retrieves a phase by ID
func GetPhase(database *db.DB, id int64) (*model.Phase, error) {
	row := database.QueryRow(
		`SELECT id, project_id, name, description, order_num, status, created_at
		 FROM phases WHERE id = ?`, id,
	)
	var p model.Phase
	var idInt int64
	var createdAt string
	var description sql.NullString
	err := row.Scan(&idInt, &p.ProjectID, &p.Name, &description, &p.OrderNum, &p.Status, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get phase: %w", err)
	}
	p.ID = strconv.FormatInt(idInt, 10)
	p.Description = description.String
	p.CreatedAt, _ = db.ParseTime(createdAt)
	return &p, nil
}

// PhaseListItem represents a phase in list view
type PhaseListItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OrderNum    int    `json:"order_num"`
	Status      string `json:"status"`
	TasksTotal  int    `json:"tasks_total"`
	TasksDone   int    `json:"tasks_done"`
}

// ListPhases lists all phases for a project
func ListPhases(database *db.DB, projectID string) ([]PhaseListItem, error) {
	rows, err := database.Query(
		`SELECT p.id, p.name, p.description, p.order_num, p.status,
		        (SELECT COUNT(*) FROM tasks WHERE phase_id = p.id) as tasks_total,
		        (SELECT COUNT(*) FROM tasks WHERE phase_id = p.id AND status = 'done') as tasks_done
		 FROM phases p
		 WHERE p.project_id = ?
		 ORDER BY p.order_num`, projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}
	defer rows.Close()

	var phases []PhaseListItem
	for rows.Next() {
		var p PhaseListItem
		var description sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &description, &p.OrderNum, &p.Status, &p.TasksTotal, &p.TasksDone); err != nil {
			return nil, fmt.Errorf("scan phase: %w", err)
		}
		p.Description = description.String
		phases = append(phases, p)
	}
	return phases, nil
}

// UpdatePhaseStatus updates the status of a phase
func UpdatePhaseStatus(database *db.DB, id int64, status string) error {
	_, err := database.Exec(`UPDATE phases SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("update phase status: %w", err)
	}
	return nil
}

// StartPhase starts a phase (pending -> active)
func StartPhase(database *db.DB, id int64) error {
	phase, err := GetPhase(database, id)
	if err != nil {
		return fmt.Errorf("get phase: %w", err)
	}
	if phase.Status != "pending" {
		return fmt.Errorf("phase status must be 'pending' to start, current: %s", phase.Status)
	}
	return UpdatePhaseStatus(database, id, "active")
}

// CompletePhase completes a phase (active -> done)
func CompletePhase(database *db.DB, id int64) error {
	phase, err := GetPhase(database, id)
	if err != nil {
		return fmt.Errorf("get phase: %w", err)
	}
	if phase.Status != "active" {
		return fmt.Errorf("phase status must be 'active' to complete, current: %s", phase.Status)
	}
	return UpdatePhaseStatus(database, id, "done")
}
