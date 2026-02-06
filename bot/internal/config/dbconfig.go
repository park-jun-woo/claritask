package config

import (
	"fmt"
	"strings"

	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/pagination"
)

// DBConfig represents a key-value config entry in global DB
type DBConfig struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updated_at"`
}

// GetDBConfig retrieves a single config value by key
func GetDBConfig(key string) types.Result {
	if key == "" {
		return types.Result{Success: false, Message: "키를 입력하세요"}
	}

	gdb, err := db.OpenGlobal()
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("DB 열기 실패: %v", err)}
	}
	defer gdb.Close()

	var cfg DBConfig
	err = gdb.QueryRow("SELECT key, value, updated_at FROM config WHERE key = ?", key).
		Scan(&cfg.Key, &cfg.Value, &cfg.UpdatedAt)
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("설정 '%s' 없음", key)}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("⚙️ %s = %s", cfg.Key, cfg.Value),
		Data:    &cfg,
	}
}

// SetDBConfig upserts a config key-value pair
func SetDBConfig(key, value string) types.Result {
	if key == "" {
		return types.Result{Success: false, Message: "키를 입력하세요"}
	}
	if value == "" {
		return types.Result{Success: false, Message: "값을 입력하세요"}
	}

	gdb, err := db.OpenGlobal()
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("DB 열기 실패: %v", err)}
	}
	defer gdb.Close()

	now := db.TimeNow()
	_, err = gdb.Exec(
		"INSERT OR REPLACE INTO config (key, value, updated_at) VALUES (?, ?, ?)",
		key, value, now,
	)
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("설정 저장 실패: %v", err)}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("✅ %s = %s", key, value),
	}
}

// ListDBConfig returns all config entries with pagination
func ListDBConfig(req pagination.PageRequest) types.Result {
	gdb, err := db.OpenGlobal()
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("DB 열기 실패: %v", err)}
	}
	defer gdb.Close()

	var total int
	if err := gdb.QueryRow("SELECT COUNT(*) FROM config").Scan(&total); err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("조회 실패: %v", err)}
	}

	rows, err := gdb.Query(
		"SELECT key, value, updated_at FROM config ORDER BY key LIMIT ? OFFSET ?",
		req.Limit(), req.Offset(),
	)
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("조회 실패: %v", err)}
	}
	defer rows.Close()

	var items []DBConfig
	for rows.Next() {
		var cfg DBConfig
		if err := rows.Scan(&cfg.Key, &cfg.Value, &cfg.UpdatedAt); err != nil {
			continue
		}
		items = append(items, cfg)
	}

	pageResp := pagination.NewPageResponse(items, req.Page, req.PageSize, total)

	if total == 0 {
		return types.Result{
			Success: true,
			Message: "설정이 없습니다",
			Data:    &pageResp,
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚙️ 설정 목록 (%d건)\n", total))
	for _, cfg := range items {
		sb.WriteString(fmt.Sprintf("  %s = %s\n", cfg.Key, cfg.Value))
	}

	if pageResp.HasNext {
		sb.WriteString(fmt.Sprintf("\n[다음:config list -p %d]", pageResp.Page+1))
	}

	return types.Result{
		Success: true,
		Message: sb.String(),
		Data:    &pageResp,
	}
}

// DeleteDBConfig deletes a config entry by key
func DeleteDBConfig(key string, confirmed bool) types.Result {
	if key == "" {
		return types.Result{Success: false, Message: "키를 입력하세요"}
	}

	gdb, err := db.OpenGlobal()
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("DB 열기 실패: %v", err)}
	}
	defer gdb.Close()

	// Check existence
	var value string
	err = gdb.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("설정 '%s' 없음", key)}
	}

	if !confirmed {
		return types.Result{
			Success:    true,
			Message:    fmt.Sprintf("설정 '%s' (값: %s)을 삭제하시겠습니까?\n[확인:config delete %s yes] [취소:config delete %s no]", key, value, key, key),
			NeedsInput: true,
			Context:    fmt.Sprintf("config delete %s", key),
		}
	}

	_, err = gdb.Exec("DELETE FROM config WHERE key = ?", key)
	if err != nil {
		return types.Result{Success: false, Message: fmt.Sprintf("삭제 실패: %v", err)}
	}

	return types.Result{
		Success: true,
		Message: fmt.Sprintf("✅ 설정 '%s' 삭제됨", key),
	}
}
