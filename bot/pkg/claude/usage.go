package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StatsCache represents the structure of ~/.claude/stats-cache.json
type StatsCache struct {
	Version            int                `json:"version"`
	LastComputedDate   string             `json:"lastComputedDate"`
	DailyActivity      []DailyActivity    `json:"dailyActivity"`
	DailyModelTokens   []DailyModelTokens `json:"dailyModelTokens"`
	ModelUsage         map[string]*ModelUsage `json:"modelUsage"`
	TotalSessions      int                `json:"totalSessions"`
	TotalMessages      int                `json:"totalMessages"`
	FirstSessionDate   string             `json:"firstSessionDate"`
}

// DailyActivity represents a single day's activity
type DailyActivity struct {
	Date         string `json:"date"`
	MessageCount int    `json:"messageCount"`
	SessionCount int    `json:"sessionCount"`
	ToolCallCount int   `json:"toolCallCount"`
}

// DailyModelTokens represents a single day's token usage by model
type DailyModelTokens struct {
	Date          string         `json:"date"`
	TokensByModel map[string]int `json:"tokensByModel"`
}

// ModelUsage represents cumulative usage for a single model
type ModelUsage struct {
	InputTokens            int     `json:"inputTokens"`
	OutputTokens           int     `json:"outputTokens"`
	CacheReadInputTokens   int     `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int   `json:"cacheCreationInputTokens"`
	CostUSD                float64 `json:"costUSD"`
}

// GetUsage reads and parses ~/.claude/stats-cache.json
func GetUsage() (*StatsCache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("홈 디렉토리 조회 실패: %w", err)
	}

	path := filepath.Join(home, ".claude", "stats-cache.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("stats-cache.json 읽기 실패: %w", err)
	}

	var stats StatsCache
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("stats-cache.json 파싱 실패: %w", err)
	}

	return &stats, nil
}

// FormatUsage returns a human-readable string of usage statistics
func FormatUsage(stats *StatsCache) string {
	var sb strings.Builder

	sb.WriteString("📊 Claude Code 사용량\n\n")

	// Summary
	sb.WriteString(fmt.Sprintf("총 세션: %d\n", stats.TotalSessions))
	sb.WriteString(fmt.Sprintf("총 메시지: %d\n", stats.TotalMessages))
	sb.WriteString(fmt.Sprintf("마지막 집계: %s\n", stats.LastComputedDate))

	// Daily activity (last 7 days)
	sb.WriteString("\n📅 일별 활동:\n")
	start := 0
	if len(stats.DailyActivity) > 7 {
		start = len(stats.DailyActivity) - 7
	}
	for _, d := range stats.DailyActivity[start:] {
		sb.WriteString(fmt.Sprintf("  %s — 메시지:%d 세션:%d 툴콜:%d\n",
			d.Date, d.MessageCount, d.SessionCount, d.ToolCallCount))
	}

	// Model usage
	if len(stats.ModelUsage) > 0 {
		sb.WriteString("\n🤖 모델별 토큰:\n")
		for model, usage := range stats.ModelUsage {
			// Shorten model name
			name := shortenModelName(model)
			sb.WriteString(fmt.Sprintf("  %s\n", name))
			sb.WriteString(fmt.Sprintf("    입력: %s  출력: %s\n",
				formatTokenCount(usage.InputTokens), formatTokenCount(usage.OutputTokens)))
			sb.WriteString(fmt.Sprintf("    캐시읽기: %s  캐시생성: %s\n",
				formatTokenCount(usage.CacheReadInputTokens), formatTokenCount(usage.CacheCreationInputTokens)))
		}
	}

	// Daily tokens (last 7 days)
	sb.WriteString("\n📈 일별 토큰:\n")
	start = 0
	if len(stats.DailyModelTokens) > 7 {
		start = len(stats.DailyModelTokens) - 7
	}
	for _, d := range stats.DailyModelTokens[start:] {
		var parts []string
		for model, tokens := range d.TokensByModel {
			parts = append(parts, fmt.Sprintf("%s:%s", shortenModelName(model), formatTokenCount(tokens)))
		}
		sb.WriteString(fmt.Sprintf("  %s — %s\n", d.Date, strings.Join(parts, " ")))
	}

	return sb.String()
}

// shortenModelName shortens a Claude model identifier
func shortenModelName(model string) string {
	// "claude-opus-4-5-20251101" -> "opus-4.5"
	// "claude-opus-4-6" -> "opus-4.6"
	// "claude-sonnet-4-5-20250929" -> "sonnet-4.5"
	model = strings.TrimPrefix(model, "claude-")
	// Remove date suffix (e.g., -20251101)
	parts := strings.Split(model, "-")
	if len(parts) >= 3 {
		// Check if last part is a date (8 digits)
		last := parts[len(parts)-1]
		if len(last) == 8 {
			parts = parts[:len(parts)-1]
		}
	}
	// Join with proper formatting: opus-4-5 -> opus-4.5
	if len(parts) >= 3 {
		return parts[0] + "-" + parts[1] + "." + parts[2]
	}
	return strings.Join(parts, "-")
}

// formatTokenCount formats a token count with K/M suffixes
func formatTokenCount(count int) string {
	if count >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(count)/1_000_000_000)
	}
	if count >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(count)/1_000_000)
	}
	if count >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(count)/1_000)
	}
	return fmt.Sprintf("%d", count)
}
