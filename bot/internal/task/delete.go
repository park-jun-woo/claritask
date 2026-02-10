package task

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Delete deletes a task
func Delete(projectPath, id string, confirmed bool) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	// Get task info first
	var t Task
	err = localDB.QueryRow(`SELECT id, title, status FROM tasks WHERE id = ?`, id).Scan(&t.ID, &t.Title, &t.Status)
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

	// Ask for confirmation
	if !confirmed {
		return types.Result{
			Success: true,
			Message: fmt.Sprintf("작업 #%d '%s'을(를) 삭제하시겠습니까?\n[예:task delete %s yes][아니오:task delete %s no]", t.ID, t.Title, id, id),
		}
	}

	// Delete
	_, err = localDB.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("삭제 실패: %v", err),
		}
	}

	// 파일 삭제
	taskID, _ := strconv.Atoi(id)
	for _, f := range []string{
		TaskFilePath(projectPath, taskID),
		PlanFilePath(projectPath, taskID),
		ReportFilePath(projectPath, taskID),
		ErrorFilePath(projectPath, taskID),
	} {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			log.Printf("[Task] 파일 삭제 실패 (%s): %v", f, err)
		}
	}
	gitCommitTask(projectPath, taskID, "deleted")

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("작업 삭제됨: #%s %s", id, t.Title),
	}
}
