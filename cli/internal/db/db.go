package db

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// LatestVersion is the current schema version
const LatestVersion = 6

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

	// Enable WAL mode for concurrent read/write
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}

	// Set busy timeout for concurrent access
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
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

CREATE TABLE IF NOT EXISTS features (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    spec TEXT DEFAULT '',
    fdl TEXT DEFAULT '',
    fdl_hash TEXT DEFAULT '',
    skeleton_generated INTEGER DEFAULT 0,
    file_path TEXT DEFAULT '',
    content TEXT DEFAULT '',
    content_hash TEXT DEFAULT '',
    status TEXT DEFAULT 'pending'
        CHECK(status IN ('pending', 'active', 'done')),
    version INTEGER DEFAULT 1,
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
    feature_id INTEGER NOT NULL,
    parent_id INTEGER,
    skeleton_id INTEGER,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK(status IN ('pending', 'doing', 'done', 'failed')),
    title TEXT NOT NULL,
    level TEXT DEFAULT 'leaf'
        CHECK(level IN ('leaf', 'parent')),
    skill TEXT DEFAULT '',
    refs TEXT DEFAULT '',
    content TEXT DEFAULT '',
    target_file TEXT DEFAULT '',
    target_line INTEGER,
    target_function TEXT DEFAULT '',
    result TEXT DEFAULT '',
    error TEXT DEFAULT '',
    version INTEGER DEFAULT 1,
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    failed_at TEXT,
    FOREIGN KEY (feature_id) REFERENCES features(id),
    FOREIGN KEY (parent_id) REFERENCES tasks(id),
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

-- memos: scope is 'project', 'feature', or 'task'
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

CREATE TABLE IF NOT EXISTS experts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT DEFAULT '1.0.0',
    domain TEXT DEFAULT '',
    language TEXT DEFAULT '',
    framework TEXT DEFAULT '',
    path TEXT NOT NULL,
    description TEXT DEFAULT '',
    content TEXT DEFAULT '',
    content_hash TEXT DEFAULT '',
    content_backup TEXT DEFAULT '',
    status TEXT DEFAULT 'active'
        CHECK(status IN ('active', 'archived')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS project_experts (
    project_id TEXT NOT NULL,
    expert_id TEXT NOT NULL,
    assigned_at TEXT NOT NULL,
    PRIMARY KEY (project_id, expert_id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (expert_id) REFERENCES experts(id)
);

CREATE TABLE IF NOT EXISTS expert_assignments (
    expert_id TEXT NOT NULL,
    feature_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (expert_id, feature_id),
    FOREIGN KEY (expert_id) REFERENCES experts(id),
    FOREIGN KEY (feature_id) REFERENCES features(id)
);
`
	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Migration: add columns to existing tables (ignore if already exists)
	migrations := []string{
		"ALTER TABLE features ADD COLUMN version INTEGER DEFAULT 1",
		"ALTER TABLE tasks ADD COLUMN version INTEGER DEFAULT 1",
		"ALTER TABLE tasks ADD COLUMN parent_id INTEGER",
		"ALTER TABLE tasks ADD COLUMN level TEXT DEFAULT 'leaf'",
		"ALTER TABLE tasks ADD COLUMN skill TEXT DEFAULT ''",
		"ALTER TABLE tasks ADD COLUMN refs TEXT DEFAULT ''",
		"ALTER TABLE experts ADD COLUMN content TEXT DEFAULT ''",
		"ALTER TABLE experts ADD COLUMN content_hash TEXT DEFAULT ''",
		"ALTER TABLE experts ADD COLUMN updated_at TEXT DEFAULT ''",
		"ALTER TABLE experts ADD COLUMN description TEXT DEFAULT ''",
		"ALTER TABLE experts ADD COLUMN content_backup TEXT DEFAULT ''",
		"ALTER TABLE experts ADD COLUMN status TEXT DEFAULT 'active'",
	}
	for _, m := range migrations {
		db.Exec(m) // Ignore errors (column may already exist)
	}

	return nil
}

// TimeNow returns the current time in ISO 8601 format
func TimeNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ParseTime parses an ISO 8601 formatted string to time.Time
func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// GetVersion returns current DB version
func (db *DB) GetVersion() int {
	// Ensure _migrations table exists
	db.Exec(`CREATE TABLE IF NOT EXISTS _migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL
	)`)

	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM _migrations").Scan(&version)
	if err != nil {
		return 0
	}
	return version
}

// setVersion records migration version
func (db *DB) setVersion(version int) error {
	now := TimeNow()
	_, err := db.Exec("INSERT INTO _migrations (version, applied_at) VALUES (?, ?)", version, now)
	return err
}

// AutoMigrate runs all pending migrations
func (db *DB) AutoMigrate() error {
	// Ensure _migrations table exists
	db.Exec(`CREATE TABLE IF NOT EXISTS _migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL
	)`)

	currentVersion := db.GetVersion()

	// Run base schema migration if needed
	if currentVersion < 1 {
		if err := db.Migrate(); err != nil {
			return fmt.Errorf("base migration failed: %w", err)
		}
		if err := db.setVersion(1); err != nil {
			return fmt.Errorf("set version 1 failed: %w", err)
		}
		currentVersion = 1
	}

	// Run incremental migrations
	for version := currentVersion + 1; version <= LatestVersion; version++ {
		if err := db.runMigration(version); err != nil {
			return fmt.Errorf("migration v%d failed: %w", version, err)
		}
		if err := db.setVersion(version); err != nil {
			return fmt.Errorf("set version v%d failed: %w", version, err)
		}
	}

	return nil
}

// runMigration runs a specific migration version
func (db *DB) runMigration(version int) error {
	switch version {
	case 2:
		// Expert support - already in base schema
		return nil
	case 3:
		// Optimistic locking - already in base schema
		return nil
	case 4:
		// Expert backup fields - already in base schema
		return nil
	case 5:
		// Add indexes
		return db.migrateV5()
	case 6:
		// Add file sync fields to features
		return db.migrateV6()
	}
	return nil
}

// migrateV6 adds file sync fields to features table
func (db *DB) migrateV6() error {
	migrations := []string{
		"ALTER TABLE features ADD COLUMN file_path TEXT DEFAULT ''",
		"ALTER TABLE features ADD COLUMN content TEXT DEFAULT ''",
		"ALTER TABLE features ADD COLUMN content_hash TEXT DEFAULT ''",
	}
	for _, m := range migrations {
		db.Exec(m) // Ignore errors (column may already exist)
	}
	return nil
}

// migrateV5 adds indexes for performance
func (db *DB) migrateV5() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_features_project ON features(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_features_status ON features(status)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_feature ON tasks(feature_id)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)",
		"CREATE INDEX IF NOT EXISTS idx_task_edges_to ON task_edges(to_task_id)",
		"CREATE INDEX IF NOT EXISTS idx_feature_edges_to ON feature_edges(to_feature_id)",
		"CREATE INDEX IF NOT EXISTS idx_memos_scope ON memos(scope, scope_id)",
		"CREATE INDEX IF NOT EXISTS idx_memos_priority ON memos(priority)",
		"CREATE INDEX IF NOT EXISTS idx_skeletons_feature ON skeletons(feature_id)",
		"CREATE INDEX IF NOT EXISTS idx_skeletons_layer ON skeletons(layer)",
		"CREATE INDEX IF NOT EXISTS idx_project_experts_project ON project_experts(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_expert_assignments_feature ON expert_assignments(feature_id)",
		"CREATE INDEX IF NOT EXISTS idx_experts_status ON experts(status)",
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}
	return nil
}

// Rollback rolls back to a specific version
func (db *DB) Rollback(targetVersion int) error {
	currentVersion := db.GetVersion()
	if targetVersion >= currentVersion {
		return fmt.Errorf("target version %d must be less than current version %d", targetVersion, currentVersion)
	}

	// Remove version records
	_, err := db.Exec("DELETE FROM _migrations WHERE version > ?", targetVersion)
	if err != nil {
		return fmt.Errorf("remove migration records: %w", err)
	}

	return nil
}

// Backup creates a backup of the database file
func (db *DB) Backup() (string, error) {
	backupPath := fmt.Sprintf("%s.backup.%s", db.path, time.Now().Format("20060102150405"))

	src, err := os.Open(db.path)
	if err != nil {
		return "", fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("create backup: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("copy data: %w", err)
	}

	return backupPath, nil
}

// Path returns the database file path
func (db *DB) Path() string {
	return db.path
}
