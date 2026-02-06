package spec

import (
	"fmt"
	"strconv"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Set updates a spec field
func Set(projectPath, id, field, value string) types.Result {
	allowedFields := map[string]bool{
		"title":    true,
		"content":  true,
		"status":   true,
		"priority": true,
	}

	if !allowedFields[field] {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("허용되지 않는 필드: %s\n허용: title, content, status, priority", field),
		}
	}

	if field == "priority" {
		if _, err := strconv.Atoi(value); err != nil {
			return types.Result{
				Success: false,
				Message: "priority는 정수여야 합니다: " + value,
			}
		}
	}

	if field == "status" {
		validStatus := map[string]bool{
			"draft":      true,
			"review":     true,
			"approved":   true,
			"deprecated": true,
		}
		if !validStatus[value] {
			return types.Result{
				Success: false,
				Message: "허용되지 않는 상태: " + value + "\n허용: draft, review, approved, deprecated",
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

	var exists int
	localDB.QueryRow("SELECT COUNT(*) FROM specs WHERE id = ?", id).Scan(&exists)
	if exists == 0 {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("스펙을 찾을 수 없습니다: #%s", id),
		}
	}

	now := db.TimeNow()
	query := fmt.Sprintf("UPDATE specs SET %s = ?, updated_at = ? WHERE id = ?", field)
	_, err = localDB.Exec(query, value, now, id)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("업데이트 실패: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("스펙 #%s %s 업데이트됨\n[조회:spec get %s]", id, field, id),
	}
}
