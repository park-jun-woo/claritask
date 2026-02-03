package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"parkjunwoo.com/claritask/internal/db"
)

func TestDBOpen(t *testing.T) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "claritask-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Check file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}

func TestDBOpenCreatesDirectory(t *testing.T) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "claritask-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a path with nested directory that doesn't exist
	dbPath := filepath.Join(tmpDir, "nested", "dir", "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Check directory was created
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("database directory was not created")
	}
}

func TestDBClose(t *testing.T) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "claritask-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Close should not return error
	if err := database.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func TestDBMigrate(t *testing.T) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "claritask-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Run migration
	if err := database.Migrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Check all tables exist
	tables := []string{"projects", "features", "tasks", "context", "tech", "design", "state", "memos"}
	for _, table := range tables {
		var name string
		err := database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table '%s' does not exist: %v", table, err)
		}
	}
}

func TestDBMigrateIdempotent(t *testing.T) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "claritask-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Run migration twice - should not error
	if err := database.Migrate(); err != nil {
		t.Fatalf("first migrate failed: %v", err)
	}
	if err := database.Migrate(); err != nil {
		t.Fatalf("second migrate failed: %v", err)
	}
}

func TestTimeNow(t *testing.T) {
	t.Helper()
	now := db.TimeNow()

	// Should be ISO 8601 format (RFC3339)
	_, err := time.Parse(time.RFC3339, now)
	if err != nil {
		t.Errorf("TimeNow did not return RFC3339 format: %v", err)
	}

	// Should contain UTC timezone
	if !strings.HasSuffix(now, "Z") {
		t.Errorf("TimeNow should be UTC (end with Z), got: %s", now)
	}
}

func TestParseTime(t *testing.T) {
	t.Helper()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid RFC3339",
			input:   "2024-01-15T10:30:00Z",
			wantErr: false,
		},
		{
			name:    "valid RFC3339 with timezone",
			input:   "2024-01-15T10:30:00+09:00",
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "2024/01/15 10:30:00",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.ParseTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTime(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestParseTimeRoundTrip(t *testing.T) {
	t.Helper()
	original := db.TimeNow()
	parsed, err := db.ParseTime(original)
	if err != nil {
		t.Fatalf("failed to parse TimeNow result: %v", err)
	}

	// Format back
	formatted := parsed.UTC().Format(time.RFC3339)
	if formatted != original {
		t.Errorf("round-trip failed: original=%s, formatted=%s", original, formatted)
	}
}

func TestForeignKeysEnabled(t *testing.T) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "claritask-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	var foreignKeys int
	err = database.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil {
		t.Fatalf("failed to check foreign_keys: %v", err)
	}

	if foreignKeys != 1 {
		t.Errorf("foreign_keys should be 1, got %d", foreignKeys)
	}
}

func TestDBMigrateTableSchema(t *testing.T) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "claritask-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Test inserting into each table to verify schema
	now := db.TimeNow()

	// projects table
	_, err = database.Exec("INSERT INTO projects (id, name, description, status, created_at) VALUES (?, ?, ?, ?, ?)",
		"test", "Test", "Description", "active", now)
	if err != nil {
		t.Errorf("failed to insert into projects: %v", err)
	}

	// features table
	_, err = database.Exec("INSERT INTO features (project_id, name, description, status, created_at) VALUES (?, ?, ?, ?, ?)",
		"test", "Feature 1", "Description", "pending", now)
	if err != nil {
		t.Errorf("failed to insert into features: %v", err)
	}

	// tasks table
	_, err = database.Exec(`INSERT INTO tasks (feature_id, title, content, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		1, "Task 1", "content", "pending", now)
	if err != nil {
		t.Errorf("failed to insert into tasks: %v", err)
	}

	// context table (singleton)
	_, err = database.Exec("INSERT INTO context (id, data, created_at, updated_at) VALUES (1, ?, ?, ?)",
		"{}", now, now)
	if err != nil {
		t.Errorf("failed to insert into context: %v", err)
	}

	// tech table (singleton)
	_, err = database.Exec("INSERT INTO tech (id, data, created_at, updated_at) VALUES (1, ?, ?, ?)",
		"{}", now, now)
	if err != nil {
		t.Errorf("failed to insert into tech: %v", err)
	}

	// design table (singleton)
	_, err = database.Exec("INSERT INTO design (id, data, created_at, updated_at) VALUES (1, ?, ?, ?)",
		"{}", now, now)
	if err != nil {
		t.Errorf("failed to insert into design: %v", err)
	}

	// state table
	_, err = database.Exec("INSERT INTO state (key, value) VALUES (?, ?)",
		"test_key", "test_value")
	if err != nil {
		t.Errorf("failed to insert into state: %v", err)
	}

	// memos table
	_, err = database.Exec("INSERT INTO memos (scope, scope_id, key, data, priority, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"project", "test", "notes", "{}", 2, now, now)
	if err != nil {
		t.Errorf("failed to insert into memos: %v", err)
	}
}
