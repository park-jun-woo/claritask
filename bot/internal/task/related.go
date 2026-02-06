package task

import (
	"parkjunwoo.com/claribot/internal/db"
)

// GetRelated returns all tasks related to the given task ID
// Related tasks include:
// - Edge 연결된 Task (양방향)
// - Parent Task
// - Child Tasks
func GetRelated(localDB *db.DB, taskID int) ([]Task, error) {
	query := `
		SELECT DISTINCT t.id, t.parent_id, t.title, t.spec, t.plan, t.report, t.status, t.error, t.created_at, t.updated_at
		FROM tasks t
		WHERE t.id != ?
		AND (
			-- Edge 연결 (양방향)
			t.id IN (SELECT to_task_id FROM task_edges WHERE from_task_id = ?)
			OR t.id IN (SELECT from_task_id FROM task_edges WHERE to_task_id = ?)
			-- Parent
			OR t.id = (SELECT parent_id FROM tasks WHERE id = ?)
			-- Children
			OR t.parent_id = ?
		)
		ORDER BY t.id ASC
	`

	rows, err := localDB.Query(query, taskID, taskID, taskID, taskID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.ParentID, &t.Title, &t.Spec, &t.Plan, &t.Report, &t.Status, &t.Error, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

