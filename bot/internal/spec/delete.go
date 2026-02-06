package spec

import (
	"database/sql"
	"fmt"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
)

// Delete deletes a spec
func Delete(projectPath, id string, confirmed bool) types.Result {
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("DB 열기 실패: %v", err),
		}
	}
	defer localDB.Close()

	var s Spec
	err = localDB.QueryRow(`SELECT id, title, status FROM specs WHERE id = ?`, id).Scan(&s.ID, &s.Title, &s.Status)
	if err == sql.ErrNoRows {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("스펙을 찾을 수 없습니다: #%s", id),
		}
	}
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("조회 실패: %v", err),
		}
	}

	if !confirmed {
		return types.Result{
			Success: true,
			Message: fmt.Sprintf("스펙 #%d '%s'을(를) 삭제하시겠습니까?\n[예:spec delete %s yes][아니오:spec delete %s no]", s.ID, s.Title, id, id),
		}
	}

	_, err = localDB.Exec("DELETE FROM specs WHERE id = ?", id)
	if err != nil {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("삭제 실패: %v", err),
		}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("스펙 삭제됨: #%s %s\n[목록:spec list]", id, s.Title),
	}
}
