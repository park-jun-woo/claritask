package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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
func OpenLocal(projectPath string) (*DB, error) {
	path := filepath.Join(projectPath, ".claribot", "db.clt")
	return Open(path)
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

// MigrateGlobal creates global DB schema (projects)
func (db *DB) MigrateGlobal() error {
	schema := `
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    type TEXT DEFAULT 'dev.platform'
        CHECK(type IN ('dev.platform', 'dev.cli', 'write.webnovel')),
    description TEXT DEFAULT '',
    status TEXT DEFAULT 'active'
        CHECK(status IN ('active', 'archived')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_type ON projects(type);
`
	_, err := db.Exec(schema)
	return err
}

// MigrateLocal creates local DB schema (tasks, task_edges)
func (db *DB) MigrateLocal() error {
	schema := `
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER,
    source TEXT DEFAULT ''
        CHECK(source IN ('', 'telegram', 'cli', 'agent')),
    title TEXT NOT NULL,
    content TEXT DEFAULT '',
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'running', 'done', 'failed')),
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS task_edges (
    from_task_id INTEGER NOT NULL,
    to_task_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (from_task_id, to_task_id),
    FOREIGN KEY (from_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (to_task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_task_edges_to ON task_edges(to_task_id);

CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    source TEXT DEFAULT ''
        CHECK(source IN ('', 'telegram', 'cli')),
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'processing', 'done', 'failed')),
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    completed_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
`
	_, err := db.Exec(schema)
	return err
}
