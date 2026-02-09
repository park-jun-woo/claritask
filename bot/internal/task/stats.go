package task

import (
	"parkjunwoo.com/claribot/internal/db"
)

// GetStats returns task statistics for a project
func GetStats(projectPath string) (*Stats, error) {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return nil, err
	}
	defer localDB.Close()

	// Fix is_leaf inconsistency: tasks with no children should be leaf
	_, _ = localDB.Exec(`
		UPDATE tasks SET is_leaf = 1
		WHERE is_leaf = 0
		  AND id NOT IN (SELECT DISTINCT parent_id FROM tasks WHERE parent_id IS NOT NULL)
	`)

	stats := &Stats{}

	err = localDB.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN is_leaf = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN is_leaf = 1 AND status = 'todo' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN is_leaf = 1 AND status = 'planned' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN is_leaf = 1 AND status = 'done' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN is_leaf = 1 AND status = 'failed' THEN 1 ELSE 0 END), 0)
		FROM tasks
	`).Scan(&stats.Total, &stats.Leaf, &stats.Todo, &stats.Planned, &stats.Done, &stats.Failed)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
