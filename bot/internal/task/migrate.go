package task

import (
	"fmt"
	"log"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
)

// MigrateContentToFiles migrates content columns (spec, plan, report, error) from DB to files.
// Only creates files that don't already exist (existing files take priority).
// Returns the number of tasks that had files created.
func MigrateContentToFiles(projectPath string) (int, error) {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return 0, fmt.Errorf("DB 열기 실패: %w", err)
	}
	defer localDB.Close()

	// Check if content columns still exist
	var taskSQL string
	localDB.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='tasks'`).Scan(&taskSQL)
	if !strings.Contains(taskSQL, "spec TEXT") {
		return 0, nil // Already migrated, content columns removed
	}

	rows, err := localDB.Query(`
		SELECT id, parent_id, title, spec, plan, report, error, status, priority
		FROM tasks ORDER BY id ASC
	`)
	if err != nil {
		return 0, fmt.Errorf("tasks 조회 실패: %w", err)
	}
	defer rows.Close()

	if err := EnsureTaskDir(projectPath); err != nil {
		return 0, err
	}

	migrated := 0

	for rows.Next() {
		var (
			id       int
			parentID *int
			title    string
			spec     string
			plan     string
			report   string
			errStr   string
			status   string
			priority int
		)
		if err := rows.Scan(&id, &parentID, &title, &spec, &plan, &report, &errStr, &status, &priority); err != nil {
			log.Printf("[Migrate] scan 실패 (skipping): %v", err)
			continue
		}

		created := false

		// Task file (.md) - frontmatter + title + spec
		if _, err := ReadTaskContent(projectPath, id); err != nil {
			fm := Frontmatter{Status: status, Parent: parentID, Priority: priority}
			if err := WriteTaskContent(projectPath, id, fm, title, spec); err != nil {
				log.Printf("[Migrate] task 파일 생성 실패 (#%d): %v", id, err)
			} else {
				created = true
			}
		}

		// Plan file
		if plan != "" {
			if existing, _ := ReadPlanContent(projectPath, id); existing == "" {
				if err := WritePlanContent(projectPath, id, plan); err != nil {
					log.Printf("[Migrate] plan 파일 생성 실패 (#%d): %v", id, err)
				} else {
					created = true
				}
			}
		}

		// Report file
		if report != "" {
			if existing, _ := ReadReportContent(projectPath, id); existing == "" {
				if err := WriteReportContent(projectPath, id, report); err != nil {
					log.Printf("[Migrate] report 파일 생성 실패 (#%d): %v", id, err)
				} else {
					created = true
				}
			}
		}

		// Error file
		if errStr != "" {
			if existing, _ := ReadErrorContent(projectPath, id); existing == "" {
				if err := WriteErrorContent(projectPath, id, errStr); err != nil {
					log.Printf("[Migrate] error 파일 생성 실패 (#%d): %v", id, err)
				} else {
					created = true
				}
			}
		}

		if created {
			migrated++
		}
	}

	if err := rows.Err(); err != nil {
		return migrated, fmt.Errorf("행 순회 오류: %w", err)
	}

	// Git commit
	if migrated > 0 {
		gitCommitBatch(projectPath, fmt.Sprintf("migrate: DB → files (%d tasks)", migrated))
	}

	return migrated, nil
}
