package task

import (
	"fmt"
	"log"
	"path/filepath"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// SyncResult contains the outcome of a sync operation.
type SyncResult struct {
	Inserted int
	Updated  int
	Deleted  int
	Restored int
	Skipped  int
	Warnings []string
}

type dbTask struct {
	ID       int
	ParentID *int
	Title    string
	Status   string
	Priority int
}

// Sync synchronizes task files with the database.
// Files are the source of truth — DB is updated to match.
func Sync(projectPath string) (*SyncResult, error) {
	result := &SyncResult{}

	// 1. Scan task files
	fileMap, err := ScanTaskFiles(projectPath)
	if err != nil {
		return nil, fmt.Errorf("파일 스캔 실패: %w", err)
	}

	// 2. Load DB tasks
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return nil, fmt.Errorf("DB 열기 실패: %w", err)
	}
	defer localDB.Close()

	rows, err := localDB.Query(`SELECT id, parent_id, title, status, priority FROM tasks ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("tasks 조회 실패: %w", err)
	}

	dbMap := make(map[int]dbTask)
	for rows.Next() {
		var t dbTask
		if err := rows.Scan(&t.ID, &t.ParentID, &t.Title, &t.Status, &t.Priority); err != nil {
			log.Printf("[Sync] scan 실패: %v", err)
			continue
		}
		dbMap[t.ID] = t
	}
	rows.Close()

	// 3. Begin transaction
	tx, err := localDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}

	parentMap := make(map[int]*int)

	// Case A: file exists, DB missing → INSERT
	for id, filePath := range fileMap {
		if err := ValidateTaskFile(filePath); err != nil {
			result.Skipped++
			result.Warnings = append(result.Warnings, fmt.Sprintf("#%d: 무효한 파일 skip: %v", id, err))
			continue
		}
		tc, err := ReadTaskContent(projectPath, id)
		if err != nil {
			result.Skipped++
			result.Warnings = append(result.Warnings, fmt.Sprintf("#%d: 파일 읽기 실패: %v", id, err))
			continue
		}
		parentMap[id] = tc.Frontmatter.Parent

		if _, exists := dbMap[id]; !exists {
			// INSERT
			now := db.TimeNow()
			_, err := tx.Exec(
				`INSERT INTO tasks (id, parent_id, title, status, priority, is_leaf, depth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, 0, ?, ?)`,
				id, tc.Frontmatter.Parent, tc.Title, tc.Frontmatter.Status, tc.Frontmatter.Priority, now, now,
			)
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("#%d: INSERT 실패: %v", id, err))
			} else {
				result.Inserted++
			}
		} else {
			// Case C: both exist → check for diff, update if needed
			dbt := dbMap[id]
			needsUpdate := false
			if dbt.Title != tc.Title {
				needsUpdate = true
			}
			if dbt.Status != tc.Frontmatter.Status {
				needsUpdate = true
			}
			if dbt.Priority != tc.Frontmatter.Priority {
				needsUpdate = true
			}
			// Compare parent_id
			if (dbt.ParentID == nil) != (tc.Frontmatter.Parent == nil) {
				needsUpdate = true
			} else if dbt.ParentID != nil && tc.Frontmatter.Parent != nil && *dbt.ParentID != *tc.Frontmatter.Parent {
				needsUpdate = true
			}

			if needsUpdate {
				now := db.TimeNow()
				_, err := tx.Exec(
					`UPDATE tasks SET parent_id=?, title=?, status=?, priority=?, updated_at=? WHERE id=?`,
					tc.Frontmatter.Parent, tc.Title, tc.Frontmatter.Status, tc.Frontmatter.Priority, now, id,
				)
				if err != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("#%d: UPDATE 실패: %v", id, err))
				} else {
					result.Updated++
				}
			}
		}
	}

	// Case B: DB exists, file missing → git restore or delete
	for id := range dbMap {
		if _, exists := fileMap[id]; exists {
			continue
		}
		// Try git restore
		relPath := filepath.Join(".claribot", taskDirName, fmt.Sprintf("%d.md", id))
		if err := GitRestore(projectPath, relPath); err == nil {
			// Re-read restored file to update parentMap
			if tc, err := ReadTaskContent(projectPath, id); err == nil {
				parentMap[id] = tc.Frontmatter.Parent
			}
			result.Restored++
			log.Printf("[Sync] git restore 성공 (#%d)", id)
		} else {
			// Delete from DB
			if _, err := tx.Exec(`DELETE FROM tasks WHERE id = ?`, id); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("#%d: DELETE 실패: %v", id, err))
			} else {
				result.Deleted++
				result.Warnings = append(result.Warnings, fmt.Sprintf("#%d: 파일 없음 → DB에서 삭제됨", id))
			}
		}
	}

	// 4. Post-processing: is_leaf + depth
	tx.Exec(`UPDATE tasks SET is_leaf = 1`)
	tx.Exec(`UPDATE tasks SET is_leaf = 0 WHERE id IN (SELECT DISTINCT parent_id FROM tasks WHERE parent_id IS NOT NULL)`)

	// Rebuild parentMap for depth calculation from DB
	depthRows, err := tx.Query(`SELECT id, parent_id FROM tasks`)
	if err == nil {
		fullParentMap := make(map[int]*int)
		for depthRows.Next() {
			var id int
			var pid *int
			depthRows.Scan(&id, &pid)
			fullParentMap[id] = pid
		}
		depthRows.Close()
		for id := range fullParentMap {
			d := computeDepth(id, fullParentMap)
			tx.Exec(`UPDATE tasks SET depth = ? WHERE id = ?`, d, id)
		}
	}

	// 5. Commit
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	// 6. Git commit if changes
	changes := result.Inserted + result.Updated + result.Deleted
	if changes > 0 {
		gitCommitBatch(projectPath, fmt.Sprintf("sync: +%d ~%d -%d", result.Inserted, result.Updated, result.Deleted))
	}

	return result, nil
}

// SyncCommand handles the "task sync" command.
func SyncCommand(projectPath string) types.Result {
	result, err := Sync(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("동기화 실패: %v", err),
		}
	}

	changes := result.Inserted + result.Updated + result.Deleted
	if changes == 0 && result.Restored == 0 {
		return types.Result{
			Success: true,
			Message: "파일 ↔ DB 일치: 변경 없음",
		}
	}

	msg := fmt.Sprintf("동기화 완료: 추가 %d, 수정 %d, 삭제 %d, 복구 %d",
		result.Inserted, result.Updated, result.Deleted, result.Restored)
	if result.Skipped > 0 {
		msg += fmt.Sprintf(", 건너뜀 %d", result.Skipped)
	}
	if len(result.Warnings) > 0 {
		msg += "\n\n⚠️ 경고:"
		for _, w := range result.Warnings {
			msg += "\n  - " + w
		}
	}

	return types.Result{
		Success: true,
		Message: msg,
	}
}
