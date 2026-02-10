package task

import (
	"database/sql"
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Get gets task details
func Get(projectPath, id string) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var t Task
	err = localDB.QueryRow(`
		SELECT id, parent_id, title, status, priority, is_leaf, depth, created_at, updated_at
		FROM tasks WHERE id = ?
	`, id).Scan(&t.ID, &t.ParentID, &t.Title, &t.Status, &t.Priority, &t.IsLeaf, &t.Depth, &t.CreatedAt, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업을 찾을 수 없습니다: #%s", id),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}

	// 삭제 보호: DB에 있는데 파일 없으면 git restore 시도
	CheckAndRestoreTaskFile(projectPath, t.ID)

	// Load content from files (sole source of truth)
	LoadContent(projectPath, &t)

	statusIcon := statusToIcon(t.Status)
	msg := fmt.Sprintf("%s #%d %s\nStatus: %s\nCreated: %s", statusIcon, t.ID, t.Title, t.Status, t.CreatedAt)
	if t.Priority != 0 {
		msg += fmt.Sprintf("\nPriority: %d", t.Priority)
	}
	if t.Spec != "" {
		msg += fmt.Sprintf("\n\n📝 Spec:\n%s", t.Spec)
	}
	if t.Plan != "" {
		msg += fmt.Sprintf("\n\n📋 Plan:\n%s", t.Plan)
	}
	if t.Report != "" {
		msg += fmt.Sprintf("\n\n📄 Report:\n%s", t.Report)
	}
	if t.Error != "" {
		msg += fmt.Sprintf("\n\n❌ Error:\n%s", t.Error)
	}

	// Add action buttons based on status
	switch t.Status {
	case "todo":
		msg += fmt.Sprintf("\n[Plan 생성:task plan %d][삭제:task delete %d]", t.ID, t.ID)
	case "planned":
		msg += fmt.Sprintf("\n[실행:task run %d][삭제:task delete %d]", t.ID, t.ID)
	case "done", "failed":
		msg += fmt.Sprintf("\n[삭제:task delete %d]", t.ID)
	}

	return types.Result{
		Success: true,
		Message: msg,
		Data:    &t,
	}
}
