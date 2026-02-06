package config

import (
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.setDefaults()

	if cfg.Service.Host != DefaultHost {
		t.Errorf("Host = %q, want %q", cfg.Service.Host, DefaultHost)
	}
	if cfg.Service.Port != DefaultPort {
		t.Errorf("Port = %d, want %d", cfg.Service.Port, DefaultPort)
	}
	if cfg.Claude.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %d, want %d", cfg.Claude.Timeout, DefaultTimeout)
	}
	if cfg.Claude.MaxTimeout != DefaultMaxTimeout {
		t.Errorf("MaxTimeout = %d, want %d", cfg.Claude.MaxTimeout, DefaultMaxTimeout)
	}
	if cfg.Claude.Max != DefaultMaxClaude {
		t.Errorf("Max = %d, want %d", cfg.Claude.Max, DefaultMaxClaude)
	}
	if cfg.Pagination.PageSize != DefaultPageSize {
		t.Errorf("PageSize = %d, want %d", cfg.Pagination.PageSize, DefaultPageSize)
	}
	if cfg.Log.Level != DefaultLogLevel {
		t.Errorf("LogLevel = %q, want %q", cfg.Log.Level, DefaultLogLevel)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name         string
		cfg          Config
		wantWarnings int
	}{
		{
			name: "valid config",
			cfg: Config{
				Service:    ServiceConfig{Host: "127.0.0.1", Port: 8080},
				Claude:     ClaudeConfig{Timeout: 600, MaxTimeout: 1800, Max: 3},
				Pagination: PaginationConfig{PageSize: 10},
			},
			wantWarnings: 0,
		},
		{
			name: "invalid port",
			cfg: Config{
				Service:    ServiceConfig{Host: "127.0.0.1", Port: 99999},
				Claude:     ClaudeConfig{Timeout: 600, MaxTimeout: 1800, Max: 3},
				Pagination: PaginationConfig{PageSize: 10},
			},
			wantWarnings: 1,
		},
		{
			name: "timeout too low",
			cfg: Config{
				Service:    ServiceConfig{Host: "127.0.0.1", Port: 8080},
				Claude:     ClaudeConfig{Timeout: 10, MaxTimeout: 1800, Max: 3},
				Pagination: PaginationConfig{PageSize: 10},
			},
			wantWarnings: 1,
		},
		{
			name: "max claude too high",
			cfg: Config{
				Service:    ServiceConfig{Host: "127.0.0.1", Port: 8080},
				Claude:     ClaudeConfig{Timeout: 600, MaxTimeout: 1800, Max: 100},
				Pagination: PaginationConfig{PageSize: 10},
			},
			wantWarnings: 1,
		},
		{
			name: "page size too high",
			cfg: Config{
				Service:    ServiceConfig{Host: "127.0.0.1", Port: 8080},
				Claude:     ClaudeConfig{Timeout: 600, MaxTimeout: 1800, Max: 3},
				Pagination: PaginationConfig{PageSize: 500},
			},
			wantWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := tt.cfg.Validate()
			if len(warnings) != tt.wantWarnings {
				t.Errorf("Validate() returned %d warnings, want %d: %v", len(warnings), tt.wantWarnings, warnings)
			}
		})
	}
}

func TestIsTelegramEnabled(t *testing.T) {
	tests := []struct {
		token    string
		expected bool
	}{
		{"", false},
		{"BOT_TOKEN", false},
		{"123456:ABC-DEF", true},
	}

	for _, tt := range tests {
		cfg := &Config{}
		cfg.Telegram.Token = tt.token
		result := cfg.IsTelegramEnabled()
		if result != tt.expected {
			t.Errorf("IsTelegramEnabled(%q) = %v, want %v", tt.token, result, tt.expected)
		}
	}
}

func TestGetLogFilePath(t *testing.T) {
	cfg := &Config{}

	// Empty path
	cfg.Log.File = ""
	if cfg.GetLogFilePath() != "" {
		t.Error("Empty file should return empty path")
	}

	// Absolute path
	cfg.Log.File = "/var/log/test.log"
	if cfg.GetLogFilePath() != "/var/log/test.log" {
		t.Errorf("Absolute path mismatch: %s", cfg.GetLogFilePath())
	}

	// Home path expansion
	cfg.Log.File = "~/test.log"
	path := cfg.GetLogFilePath()
	if path == "~/test.log" {
		t.Error("~ should be expanded")
	}
}
