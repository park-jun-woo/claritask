package task

import (
	"fmt"
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

	// Update field and updated_at
	query := fmt.Sprintf("UPDATE tasks SET %s = ?, updated_at = ? WHERE id = ?", field)
	_, err = localDB.Exec(query, value, now, id)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("업데이트 실패: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("작업 #%s %s 업데이트됨\n[조회:task get %s]", id, field, id),
	}
}
