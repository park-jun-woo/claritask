package project

import (
	"database/sql"
	"fmt"
	"strconv"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// DefaultParallel is the default parallel execution count
const DefaultParallel = 3

// Get gets project details
func Get(id string) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to open global db: %v", err),
		}
	}
	defer globalDB.Close()

	var p Project
	var pinned int
	err = globalDB.QueryRow(`
		SELECT id, name, path, description, status, category, pinned, last_accessed, created_at, updated_at
		FROM projects WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Path, &p.Description, &p.Status, &p.Category, &pinned, &p.LastAccessed, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("project not found: %s", id),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("failed to query project: %v", err),
		}
	}
	p.Pinned = pinned == 1

	// Read parallel from local DB config
	p.Parallel = DefaultParallel
	localDB, err := db.OpenLocal(p.Path)
	if err == nil {
		defer localDB.Close()
		var val string
		err = localDB.QueryRow("SELECT value FROM config WHERE key = 'parallel'").Scan(&val)
		if err == nil {
			if n, err := strconv.Atoi(val); err == nil && n >= 1 {
				p.Parallel = n
			}
		}
	}

	msg := fmt.Sprintf("Project: %s\nPath: %s\nDescription: %s\nStatus: %s\nParallel: %d\nCreated: %s",
		p.ID, p.Path, p.Description, p.Status, p.Parallel, p.CreatedAt)

	return types.Result{
		Success: true,
		Message: msg,
		Data:    &p,
	}
}
