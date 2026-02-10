package task

import (
	"fmt"
	"log"
	"strconv"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Set updates a task field
func Set(projectPath, id, field, value string) types.Result {
	// Allowed fields
	allowedFields := map[string]bool{
		"title":    true,
		"spec":     true,
		"plan":     true,
		"report":   true,
		"status":   true,
		"priority": true,
	}

	if !allowedFields[field] {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("허용되지 않는 필드: %s\n허용: title, spec, plan, report, status, priority", field),
		}
	}

	// Validate priority (must be integer)
	if field == "priority" {
		if _, err := strconv.Atoi(value); err != nil {
			return types.Result{
				Success: false,
				Message: "priority는 정수여야 합니다: " + value,
			}
		}
	}

	// Validate status values
	if field == "status" {
		validStatus := map[string]bool{
			"todo":    true,
			"planned": true,
			"split":   true,
			"done":    true,
			"failed":  true,
		}
		if !validStatus[value] {
			return types.Result{
				Success: false,
				Message: "허용되지 않는 상태: " + value + "\n허용: todo, planned, split, done, failed",
			}
		}
	}

	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	// Check if task exists and get current status
	var currentStatus string
	err = localDB.QueryRow("SELECT status FROM tasks WHERE id = ?", id).Scan(&currentStatus)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업을 찾을 수 없습니다: #%s", id),
		}
	}

	// Prevent changing from 'split' status (has children, cannot revert)
	if field == "status" && currentStatus == "split" && value != "split" {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업 #%s는 하위 작업이 있어 split 상태를 변경할 수 없습니다", id),
		}
	}

	now := db.TimeNow()
	taskID, _ := strconv.Atoi(id)

	switch field {
	case "spec":
		tc, err := ReadTaskContent(projectPath, taskID)
		if err != nil {
			// File doesn't exist — query DB for meta and create new file
			var parentID *int
			var title string
			localDB.QueryRow("SELECT parent_id, title FROM tasks WHERE id = ?", id).Scan(&parentID, &title)
			fm := Frontmatter{Status: currentStatus, Parent: parentID}
			if err := WriteTaskContent(projectPath, taskID, fm, title, value); err != nil {
				log.Printf("[Task] spec 파일 생성 실패 (#%d): %v", taskID, err)
			}
		} else {
			if err := WriteTaskContent(projectPath, taskID, tc.Frontmatter, tc.Title, value); err != nil {
				log.Printf("[Task] spec 파일 업데이트 실패 (#%d): %v", taskID, err)
			}
		}
		localDB.Exec("UPDATE tasks SET updated_at = ? WHERE id = ?", now, id)
		gitCommitTask(projectPath, taskID, "spec updated")

	case "plan":
		if err := WritePlanContent(projectPath, taskID, value); err != nil {
			log.Printf("[Task] plan 파일 생성 실패 (#%d): %v", taskID, err)
		}
		localDB.Exec("UPDATE tasks SET updated_at = ? WHERE id = ?", now, id)
		gitCommitTask(projectPath, taskID, "plan updated")

	case "report":
		if err := WriteReportContent(projectPath, taskID, value); err != nil {
			log.Printf("[Task] report 파일 생성 실패 (#%d): %v", taskID, err)
		}
		localDB.Exec("UPDATE tasks SET updated_at = ? WHERE id = ?", now, id)
		gitCommitTask(projectPath, taskID, "report updated")

	case "title":
		// DB update + file title sync
		localDB.Exec("UPDATE tasks SET title = ?, updated_at = ? WHERE id = ?", value, now, id)
		if tc, err := ReadTaskContent(projectPath, taskID); err == nil {
			if err := WriteTaskContent(projectPath, taskID, tc.Frontmatter, value, tc.Body); err != nil {
				log.Printf("[Task] title 파일 업데이트 실패 (#%d): %v", taskID, err)
			}
		}
		gitCommitTask(projectPath, taskID, "title updated")

	case "status":
		query := "UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?"
		_, err = localDB.Exec(query, value, now, id)
		if err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("업데이트 실패: %v", err),
			}
		}
		updateTaskFileStatus(projectPath, taskID, value)
		gitCommitTask(projectPath, taskID, value)

	case "priority":
		query := "UPDATE tasks SET priority = ?, updated_at = ? WHERE id = ?"
		_, err = localDB.Exec(query, value, now, id)
		if err != nil {
			return types.Result{
				Success: false,
				Message: fmt.Sprintf("업데이트 실패: %v", err),
			}
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("작업 #%s %s 업데이트됨\n[조회:task get %s]", id, field, id),
	}
}
