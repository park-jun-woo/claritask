package task

import (
	"parkjunwoo.com/claribot/internal/db"
)

// GetRelated returns all tasks related to the given task ID
// Related tasks include:
// - Parent Task
// - Child Tasks
func GetRelated(localDB *db.DB, taskID int) ([]Task, error) {
	query := `
		SELECT DISTINCT t.id, t.parent_id, t.title, t.status, t.created_at, t.updated_at
		FROM tasks t
		WHERE t.id != ?
		AND (
			-- Parent
			t.id = (SELECT parent_id FROM tasks WHERE id = ?)
			-- Children
			OR t.parent_id = ?
		)
		ORDER BY t.id ASC
	`

	rows, err := localDB.Query(query, taskID, taskID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.ParentID, &t.Title, &t.Status, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

