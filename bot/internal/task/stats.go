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

	stats := &Stats{}

	err = localDB.QueryRow(`
		SELECT
			COUNT(*),
			SUM(CASE WHEN is_leaf = 1 THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_leaf = 1 AND status = 'spec_ready' THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_leaf = 1 AND status = 'plan_ready' THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_leaf = 1 AND status = 'done' THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_leaf = 1 AND status = 'failed' THEN 1 ELSE 0 END)
		FROM tasks
	`).Scan(&stats.Total, &stats.Leaf, &stats.SpecReady, &stats.PlanReady, &stats.Done, &stats.Failed)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
