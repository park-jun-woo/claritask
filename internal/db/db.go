package db

import (
	"database/sql"
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
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	return &DB{DB: db, path: path}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Migrate creates all tables if they don't exist
func (db *DB) Migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'active',
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS phases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    order_num INTEGER,
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'active', 'done')),
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS features (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    spec TEXT DEFAULT '',
    fdl TEXT DEFAULT '',
    fdl_hash TEXT DEFAULT '',
    skeleton_generated INTEGER DEFAULT 0,
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'active', 'done')),
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id),
    UNIQUE(project_id, name)
);

CREATE TABLE IF NOT EXISTS feature_edges (
    from_feature_id INTEGER NOT NULL,
    to_feature_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (from_feature_id, to_feature_id),
    FOREIGN KEY (from_feature_id) REFERENCES features(id),
    FOREIGN KEY (to_feature_id) REFERENCES features(id)
);

CREATE TABLE IF NOT EXISTS skeletons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feature_id INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    layer TEXT NOT NULL
        CHECK(layer IN ('model', 'service', 'api', 'ui')),
    checksum TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (feature_id) REFERENCES features(id)
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    phase_id INTEGER NOT NULL,
    parent_id INTEGER DEFAULT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK(status IN ('pending', 'doing', 'done', 'failed')),
    title TEXT NOT NULL,
    level TEXT DEFAULT ''
        CHECK(level IN ('', 'node', 'leaf')),
    skill TEXT DEFAULT '',
    "references" TEXT DEFAULT '[]',
    content TEXT DEFAULT '',
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    feature_id INTEGER DEFAULT NULL,
    skeleton_id INTEGER DEFAULT NULL,
    target_file TEXT DEFAULT '',
    target_line INTEGER,
    target_function TEXT DEFAULT '',
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    failed_at TEXT,
    FOREIGN KEY (phase_id) REFERENCES phases(id),
    FOREIGN KEY (parent_id) REFERENCES tasks(id),
    FOREIGN KEY (feature_id) REFERENCES features(id),
    FOREIGN KEY (skeleton_id) REFERENCES skeletons(id)
);

CREATE TABLE IF NOT EXISTS task_edges (
    from_task_id INTEGER NOT NULL,
    to_task_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (from_task_id, to_task_id),
    FOREIGN KEY (from_task_id) REFERENCES tasks(id),
    FOREIGN KEY (to_task_id) REFERENCES tasks(id)
);

CREATE TABLE IF NOT EXISTS context (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    data TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tech (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    data TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS design (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    data TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS state (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS memos (
    scope TEXT NOT NULL,
    scope_id TEXT NOT NULL,
    key TEXT NOT NULL,
    data TEXT NOT NULL,
    priority INTEGER DEFAULT 2
        CHECK(priority IN (1, 2, 3)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (scope, scope_id, key)
);
`
	_, err := db.Exec(schema)
	return err
}

// TimeNow returns the current time in ISO 8601 format
func TimeNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ParseTime parses an ISO 8601 formatted string to time.Time
func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
