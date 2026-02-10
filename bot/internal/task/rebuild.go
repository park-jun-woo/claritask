package task

import (
	"fmt"
	"log"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

const cleanTasksSchema = `
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER,
    title TEXT NOT NULL,
    status TEXT DEFAULT 'todo'
        CHECK(status IN ('todo', 'split', 'planned', 'done', 'failed')),
    priority INTEGER DEFAULT 0,
    is_leaf INTEGER DEFAULT 1,
    depth INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_leaf ON tasks(is_leaf);
`

// Rebuild drops and recreates the tasks table from task files.
// Returns the number of tasks inserted.
func Rebuild(projectPath string) (int, error) {
	// 1. Scan task files
	fileMap, err := ScanTaskFiles(projectPath)
	if err != nil {
		return 0, fmt.Errorf("파일 스캔 실패: %w", err)
	}

	// 2. Validate and read each file
	type taskData struct {
		ID       int
		Title    string
		Status   string
		ParentID *int
		Priority int
	}
	var tasks []taskData
	parentMap := make(map[int]*int)

	for id, filePath := range fileMap {
		if err := ValidateTaskFile(filePath); err != nil {
			log.Printf("[Rebuild] 무효한 파일 skip (#%d): %v", id, err)
			continue
		}
		tc, err := ReadTaskContent(projectPath, id)
		if err != nil {
			log.Printf("[Rebuild] 파일 읽기 실패 skip (#%d): %v", id, err)
			continue
		}
		td := taskData{
			ID:       id,
			Title:    tc.Title,
			Status:   tc.Frontmatter.Status,
			ParentID: tc.Frontmatter.Parent,
			Priority: tc.Frontmatter.Priority,
		}
		tasks = append(tasks, td)
		parentMap[id] = tc.Frontmatter.Parent
	}

	// 3. Open DB and begin transaction
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return 0, fmt.Errorf("DB 열기 실패: %w", err)
	}
	defer localDB.Close()

	localDB.Exec(`PRAGMA foreign_keys=OFF`)
	defer localDB.Exec(`PRAGMA foreign_keys=ON`)

	tx, err := localDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}

	// 4. Drop and recreate with clean schema
	if _, err := tx.Exec(`DROP TABLE IF EXISTS tasks`); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("DROP TABLE 실패: %w", err)
	}
	if _, err := tx.Exec(cleanTasksSchema); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("CREATE TABLE 실패: %w", err)
	}

	// 5. Insert tasks
	now := db.TimeNow()
	stmt, err := tx.Prepare(`INSERT INTO tasks (id, parent_id, title, status, priority, is_leaf, depth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, 0, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("INSERT 준비 실패: %w", err)
	}
	defer stmt.Close()

	for _, t := range tasks {
		if _, err := stmt.Exec(t.ID, t.ParentID, t.Title, t.Status, t.Priority, now, now); err != nil {
			log.Printf("[Rebuild] INSERT 실패 (#%d): %v", t.ID, err)
			continue
		}
	}

	// 6. Compute is_leaf
	tx.Exec(`UPDATE tasks SET is_leaf = 0 WHERE id IN (SELECT DISTINCT parent_id FROM tasks WHERE parent_id IS NOT NULL)`)

	// 7. Compute depth
	for _, t := range tasks {
		d := computeDepth(t.ID, parentMap)
		tx.Exec(`UPDATE tasks SET depth = ? WHERE id = ?`, d, t.ID)
	}

	// 8. Fix sqlite_sequence
	tx.Exec(`DELETE FROM sqlite_sequence WHERE name='tasks'`)
	tx.Exec(`INSERT INTO sqlite_sequence (name, seq) VALUES ('tasks', (SELECT COALESCE(MAX(id), 0) FROM tasks))`)

	// 9. Commit
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	// 10. Git commit
	gitCommitBatch(projectPath, fmt.Sprintf("rebuild: %d tasks from files", len(tasks)))

	return len(tasks), nil
}

// RebuildCommand handles the "task rebuild" command with confirmation.
func RebuildCommand(projectPath string, confirmed bool) types.Result {
	if !confirmed {
		return types.Result{
			Success:    true,
			Message:    "DB를 파일에서 재구축하시겠습니까? 기존 DB의 task 데이터가 모두 교체됩니다.\n[예:task rebuild yes] [아니오:task rebuild no]",
			NeedsInput: true,
			Context:    "task rebuild",
		}
	}

	count, err := Rebuild(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("재구축 실패: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("DB 재구축 완료: %d개 task 파일에서 복원됨 (content 컬럼 제거)", count),
	}
}
