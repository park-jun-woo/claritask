package task

import (
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Set updates a task field
func Set(projectPath, id, field, value string) types.Result {
	// Allowed fields
	allowedFields := map[string]bool{
		"title":   true,
		"content": true,
		"status":  true,
		"result":  true,
	}

	if !allowedFields[field] {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("허용되지 않는 필드: %s\n허용: title, content, status, result", field),
		}
	}

	// Validate status values
	if field == "status" {
		validStatus := map[string]bool{
			"pending": true,
			"running": true,
			"done":    true,
			"failed":  true,
		}
		if !validStatus[value] {
			return types.Result{
				Success: false,
				Message: "허용되지 않는 상태: " + value + "\n허용: pending, running, done, failed",
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

	query := fmt.Sprintf("UPDATE tasks SET %s = ? WHERE id = ?", field)
	result, err := localDB.Exec(query, value, id)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("업데이트 실패: %v", err),
		}
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("작업을 찾을 수 없습니다: #%s", id),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("작업 #%s %s 업데이트됨\n[조회:task get %s]", id, field, id),
	}
}
