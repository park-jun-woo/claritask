package project

import (
	"database/sql"
	"fmt"
	"strconv"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/claude"
)

// Set sets a project configuration field
func Set(id, field, value string) types.Result {
	if id == "" {
		return types.Result{Success: false, Message: "프로젝트 ID를 입력하세요"}
	}
	if field == "" {
		return types.Result{Success: false, Message: "설정할 필드를 입력하세요"}
	}

	switch field {
	case "parallel":
		return setParallel(id, value)
	case "description":
		return setDescription(id, value)
	default:
		return types.Result{Success: false, Message: fmt.Sprintf("알 수 없는 필드: %s (지원: parallel, description)", field)}
	}
}

// setDescription sets the description of a project
func setDescription(id, value string) types.Result {
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("전역 DB 열기 실패: %v", err)}
	}
	defer globalDB.Close()

	now := db.TimeNow()
	result, err := globalDB.Exec(
		"UPDATE projects SET description = ?, updated_at = ? WHERE id = ?",
		value, now, id,
	)
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("설명 저장 실패: %v", err)}
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return types.Result{Success: false, Message: fmt.Sprintf("프로젝트를 찾을 수 없습니다: %s", id)}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("✅ 프로젝트 '%s' description 업데이트됨", id),
	}
}

// setParallel sets the parallel execution count for a project
func setParallel(id, value string) types.Result {
	n, err := strconv.Atoi(value)
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("숫자를 입력하세요: %s", value)}
	}

	maxClaude := claude.GetStatus().Max
	if n < 1 || n > maxClaude {
		return types.Result{
			Success: false,
			Message: fmt.Sprintf("범위 오류: parallel은 1~%d 사이여야 합니다 (현재 claude.max=%d)", maxClaude, maxClaude),
		}
	}

	// Get project path from global DB
	globalDB, err := db.OpenGlobal()
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("전역 DB 열기 실패: %v", err)}
	}
	defer globalDB.Close()

	var projectPath string
	err = globalDB.QueryRow("SELECT path FROM projects WHERE id = ?", id).Scan(&projectPath)
	if err == sql.ErrNoRows {
		return types.Result{Success: false, Message: fmt.Sprintf("프로젝트를 찾을 수 없습니다: %s", id)}
	}
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("프로젝트 조회 실패: %v", err)}
	}

	// Open local DB and upsert config
	localDB, err := db.OpenLocal(projectPath)
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("로컬 DB 열기 실패: %v", err)}
	}
	defer localDB.Close()

	now := db.TimeNow()
	_, err = localDB.Exec(
		"INSERT OR REPLACE INTO config (key, value, updated_at) VALUES (?, ?, ?)",
		"parallel", strconv.Itoa(n), now,
	)
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("설정 저장 실패: %v", err)}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("✅ 프로젝트 '%s' parallel = %d", id, n),
	}
}
