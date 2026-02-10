package task

import (
	"fmt"
	"log"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Add adds a new task with optional parent and spec
func Add(projectPath, title string, parentID *int, spec string) types.Result {
	// Auto-generate title from spec first line if empty
	if title == "" && spec != "" {
		firstLine := strings.SplitN(spec, "\n", 2)[0]
		firstLine = strings.TrimSpace(firstLine)
		runes := []rune(firstLine)
		if len(runes) > 100 {
			firstLine = string(runes[:100])
		}
		title = firstLine
	}
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	// Calculate depth from parent
	depth := 0
	if parentID != nil {
		var parentDepth int
		err := localDB.QueryRow("SELECT depth FROM tasks WHERE id = ?", *parentID).Scan(&parentDepth)
		if err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("부모 작업을 찾을 수 없습니다: #%d", *parentID),
			}
		}
		depth = parentDepth + 1
	}

	now := db.TimeNow()
	result, err := localDB.Exec(`
		INSERT INTO tasks (parent_id, title, status, is_leaf, depth, created_at, updated_at)
		VALUES (?, ?, 'todo', 1, ?, ?, ?)
	`, parentID, title, depth, now, now)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("추가 실패: %v", err),
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 ID 획득 실패: %v", err),
		}
	}

	// Dual-write: create task file
	fm := Frontmatter{Status: "todo", Parent: parentID}
	if err := WriteTaskContent(projectPath, int(id), fm, title, spec); err != nil {
		log.Printf("[Task] task 파일 생성 실패 (#%d): %v", id, err)
	}
	gitCommitTask(projectPath, int(id), "created")

	msg := fmt.Sprintf("작업 추가됨: #%d %s", id, title)
	if parentID != nil {
		// Update parent: set status to 'split' and is_leaf to 0
		_, err = localDB.Exec(`
			UPDATE tasks SET status = 'split', is_leaf = 0, updated_at = ? WHERE id = ?
		`, now, *parentID)
		if err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("부모 작업 상태 업데이트 실패: %v", err),
			}
		}
		msg += fmt.Sprintf(" (부모: #%d → split, depth: %d)", *parentID, depth)
	}

	return types.Result{
		Success: true,
		Message: msg,
		Data: &Task{
			ID:        int(id),
			ParentID:  parentID,
			Title:     title,
			Spec:      spec,
			Status:    "todo",
			IsLeaf:    true,
			Depth:     depth,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}
