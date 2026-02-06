package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps sql.DB with additional functionality
type DB struct {
	*sql.DB
	path string
}

// Open opens a database connection, creating the file if necessary
func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, err
	}

	return &DB{DB: db, path: path}, nil
}

// OpenGlobal opens the global database (~/.claribot/db.clt)
func OpenGlobal() (*DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	path := filepath.Join(home, ".claribot", "db.clt")
	return Open(path)
}

// OpenLocal opens a project's local database (project/.claribot/db.clt)
// Automatically runs migrations to ensure schema is up to date.
func OpenLocal(projectPath string) (*DB, error) {
	path := filepath.Join(projectPath, ".claribot", "db.clt")
	db, err := Open(path)
	if err != nil {
		return nil, err
	}
	if err := db.MigrateLocal(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate local: %w", err)
	}
	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Path returns the database file path
func (db *DB) Path() string {
	return db.path
}

// TimeNow returns the current time in ISO 8601 format
func TimeNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// MigrateGlobal creates global DB schema (projects, schedules)
func (db *DB) MigrateGlobal() error {
	schema := `
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    type TEXT DEFAULT 'dev.platform'
        CHECK(type IN ('dev.platform', 'dev.cli', 'write.webnovel', 'life.family')),
    description TEXT DEFAULT '',
    status TEXT DEFAULT 'active'
        CHECK(status IN ('active', 'archived')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_type ON projects(type);

CREATE TABLE IF NOT EXISTS schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT,
    cron_expr TEXT NOT NULL,
    message TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'claude'
        CHECK(type IN ('claude', 'bash')),
    enabled INTEGER DEFAULT 1,
    run_once INTEGER DEFAULT 0,
    last_run TEXT,
    next_run TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled);
CREATE INDEX IF NOT EXISTS idx_schedules_project ON schedules(project_id);

CREATE TABLE IF NOT EXISTS schedule_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_id INTEGER NOT NULL,
    status TEXT DEFAULT 'running'
        CHECK(status IN ('running', 'done', 'failed')),
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    started_at TEXT NOT NULL,
    completed_at TEXT,
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_schedule_runs_schedule ON schedule_runs(schedule_id);
CREATE INDEX IF NOT EXISTS idx_schedule_runs_status ON schedule_runs(status);

CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT,
    content TEXT NOT NULL,
    source TEXT DEFAULT ''
        CHECK(source IN ('', 'telegram', 'cli', 'schedule')),
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'processing', 'done', 'failed')),
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    completed_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_project ON messages(project_id);

CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS auth (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
`
	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Run migrations for existing tables
	migrations := []string{
		// Add run_once column to schedules if not exists
		`ALTER TABLE schedules ADD COLUMN run_once INTEGER DEFAULT 0`,
		// Add project_id column to messages if not exists (for old local messages migrated to global)
		`ALTER TABLE messages ADD COLUMN project_id TEXT`,
		// Add type column to schedules if not exists
		`ALTER TABLE schedules ADD COLUMN type TEXT NOT NULL DEFAULT 'claude'`,
	}

	// Recreate projects table to update CHECK constraint (add life.family type)
	recreateProjects := []string{
		`CREATE TABLE IF NOT EXISTS projects_new (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			type TEXT DEFAULT 'dev.platform'
				CHECK(type IN ('dev.platform', 'dev.cli', 'write.webnovel', 'life.family')),
			description TEXT DEFAULT '',
			status TEXT DEFAULT 'active'
				CHECK(status IN ('active', 'archived')),
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`INSERT OR IGNORE INTO projects_new SELECT * FROM projects`,
		`DROP TABLE projects`,
		`ALTER TABLE projects_new RENAME TO projects`,
		`CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status)`,
		`CREATE INDEX IF NOT EXISTS idx_projects_type ON projects(type)`,
	}

	for _, migration := range migrations {
		// Ignore errors (column already exists)
		db.Exec(migration)
	}

	// Check if projects table needs CHECK constraint update
	var checkInfo string
	err = db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='projects'`).Scan(&checkInfo)
	if err == nil && !strings.Contains(checkInfo, "life.family") {
		for _, stmt := range recreateProjects {
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("projects migration failed: %w", err)
			}
		}
	}

	return nil
}

// MigrateLocal creates local DB schema (tasks, traversals, config)
func (db *DB) MigrateLocal() error {
	schema := `
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER,
    title TEXT NOT NULL,
    spec TEXT DEFAULT '',
    plan TEXT DEFAULT '',
    report TEXT DEFAULT '',
    status TEXT DEFAULT 'todo'
        CHECK(status IN ('todo', 'split', 'planned', 'done', 'failed')),
    error TEXT DEFAULT '',
    is_leaf INTEGER DEFAULT 1,
    depth INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

CREATE TABLE IF NOT EXISTS traversals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK(type IN ('plan', 'run', 'cycle')),
    target_id INTEGER,
    trigger TEXT DEFAULT '',
    status TEXT DEFAULT 'running' CHECK(status IN ('running', 'done', 'failed', 'cancelled')),
    total INTEGER DEFAULT 0,
    success INTEGER DEFAULT 0,
    failed INTEGER DEFAULT 0,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    FOREIGN KEY (target_id) REFERENCES tasks(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_traversals_type ON traversals(type);
CREATE INDEX IF NOT EXISTS idx_traversals_status ON traversals(status);

CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
`
	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Drop removed tables
	db.Exec(`DROP TABLE IF EXISTS task_edges`)

	// Run migrations for existing tables (errors ignored: column/index already exists)
	migrations := []string{
		`ALTER TABLE tasks ADD COLUMN is_leaf INTEGER DEFAULT 1`,
		`ALTER TABLE tasks ADD COLUMN depth INTEGER DEFAULT 0`,
		`ALTER TABLE tasks ADD COLUMN spec TEXT DEFAULT ''`,
		`ALTER TABLE tasks ADD COLUMN plan TEXT DEFAULT ''`,
		`ALTER TABLE tasks ADD COLUMN report TEXT DEFAULT ''`,
		`ALTER TABLE tasks ADD COLUMN error TEXT DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_leaf ON tasks(is_leaf)`,
		// Migrate status values: spec_ready→todo, plan_ready→planned, subdivided→split
		`UPDATE tasks SET status = 'todo' WHERE status = 'spec_ready'`,
		`UPDATE tasks SET status = 'planned' WHERE status = 'plan_ready'`,
		`UPDATE tasks SET status = 'split' WHERE status = 'subdivided'`,
		// Add priority column for task execution ordering
		`ALTER TABLE tasks ADD COLUMN priority INTEGER DEFAULT 0`,
	}

	for _, migration := range migrations {
		// Ignore errors (column already exists)
		db.Exec(migration)
	}

	return nil
}
