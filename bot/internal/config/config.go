package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the full configuration
type Config struct {
	Service    ServiceConfig    `yaml:"service"`
	Telegram   TelegramConfig   `yaml:"telegram"`
	Claude     ClaudeConfig     `yaml:"claude"`
	Project    ProjectConfig    `yaml:"project"`
	Pagination PaginationConfig `yaml:"pagination"`
	Log        LogConfig        `yaml:"log"`
}

// ServiceConfig for HTTP server
type ServiceConfig struct {
	Host string `yaml:"host"` // default: 127.0.0.1
	Port int    `yaml:"port"` // default: 9847
}

// TelegramConfig for telegram bot
type TelegramConfig struct {
	Token        string  `yaml:"token"`          // required for telegram
	AllowedUsers []int64 `yaml:"allowed_users"`  // empty = allow all
	AdminChatID  int64   `yaml:"admin_chat_id"`  // chat ID for schedule notifications
}

// ClaudeConfig for Claude Code execution
type ClaudeConfig struct {
	Timeout    int `yaml:"timeout"`     // idle timeout seconds, default: 1200 (20 min)
	MaxTimeout int `yaml:"max_timeout"` // absolute timeout seconds, default: 1800 (30 min)
	Max        int `yaml:"max"`         // max concurrent, default: 3
}

// ProjectConfig for project management
type ProjectConfig struct {
	Path string `yaml:"path"` // default project creation path
}

// PaginationConfig for list pagination
type PaginationConfig struct {
	PageSize int `yaml:"page_size"` // default: 10
}

// LogConfig for logging
type LogConfig struct {
	Level string `yaml:"level"` // debug, info, warn, error (default: info)
	File  string `yaml:"file"`  // log file path (empty = stdout only)
}

// Defaults
const (
	DefaultHost        = "127.0.0.1"
	DefaultPort        = 9847
	DefaultTimeout     = 1200 // 20 minutes
	DefaultMaxTimeout  = 1800 // 30 minutes
	DefaultMaxClaude   = 10
	DefaultPageSize    = 10
	DefaultLogLevel    = "info"
)

// Load loads config from ~/.claribot/config.yaml
func Load() (*Config, error) {
	cfg := &Config{}
	cfg.setDefaults()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return cfg, fmt.Errorf("get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".claribot", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // use defaults
		}
		return cfg, fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("parse config file: %w", err)
	}

	// Re-apply defaults for zero values
	cfg.applyDefaults()

	return cfg, nil
}

func (c *Config) setDefaults() {
	c.Service.Host = DefaultHost
	c.Service.Port = DefaultPort
	c.Claude.Timeout = DefaultTimeout
	c.Claude.MaxTimeout = DefaultMaxTimeout
	c.Claude.Max = DefaultMaxClaude
	c.Pagination.PageSize = DefaultPageSize
	c.Log.Level = DefaultLogLevel
}

func (c *Config) applyDefaults() {
	if c.Service.Host == "" {
		c.Service.Host = DefaultHost
	}
	if c.Service.Port == 0 {
		c.Service.Port = DefaultPort
	}
	if c.Claude.Timeout == 0 {
		c.Claude.Timeout = DefaultTimeout
	}
	if c.Claude.MaxTimeout == 0 {
		c.Claude.MaxTimeout = DefaultMaxTimeout
	}
	if c.Claude.Max == 0 {
		c.Claude.Max = DefaultMaxClaude
	}
	if c.Pagination.PageSize == 0 {
		c.Pagination.PageSize = DefaultPageSize
	}
	if c.Log.Level == "" {
		c.Log.Level = DefaultLogLevel
	}
}

// Validate checks config values
func (c *Config) Validate() []string {
	var warnings []string

	if c.Service.Port < 1 || c.Service.Port > 65535 {
		warnings = append(warnings, fmt.Sprintf("invalid port %d, using default %d", c.Service.Port, DefaultPort))
		c.Service.Port = DefaultPort
	}

	if c.Claude.Timeout < 60 {
		warnings = append(warnings, fmt.Sprintf("timeout %d too low, using minimum 60", c.Claude.Timeout))
		c.Claude.Timeout = 60
	}

	if c.Claude.MaxTimeout < 60 {
		warnings = append(warnings, fmt.Sprintf("max_timeout %d too low, using minimum 60", c.Claude.MaxTimeout))
		c.Claude.MaxTimeout = 60
	}
	if c.Claude.MaxTimeout > 7200 {
		warnings = append(warnings, fmt.Sprintf("max_timeout %d too high, using maximum 7200", c.Claude.MaxTimeout))
		c.Claude.MaxTimeout = 7200
	}

	if c.Claude.Max < 1 {
		warnings = append(warnings, fmt.Sprintf("max claude instances %d invalid, using 1", c.Claude.Max))
		c.Claude.Max = 1
	}

	if c.Claude.Max > 10 {
		warnings = append(warnings, fmt.Sprintf("max claude instances %d too high, using 10", c.Claude.Max))
		c.Claude.Max = 10
	}

	if c.Pagination.PageSize < 1 {
		warnings = append(warnings, fmt.Sprintf("page_size %d invalid, using default %d", c.Pagination.PageSize, DefaultPageSize))
		c.Pagination.PageSize = DefaultPageSize
	}

	if c.Pagination.PageSize > 100 {
		warnings = append(warnings, fmt.Sprintf("page_size %d too high, using 100", c.Pagination.PageSize))
		c.Pagination.PageSize = 100
	}

	return warnings
}

// IsTelegramEnabled returns true if telegram bot is configured
func (c *Config) IsTelegramEnabled() bool {
	return c.Telegram.Token != "" && c.Telegram.Token != "BOT_TOKEN"
}

// GetLogFilePath returns absolute log file path
func (c *Config) GetLogFilePath() string {
	if c.Log.File == "" {
		return ""
	}
	// Expand ~ to home directory
	if len(c.Log.File) > 0 && c.Log.File[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, c.Log.File[1:])
	}
	return c.Log.File
}

// ReadRaw reads the raw config.yaml content
func ReadRaw() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".claribot", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read config file: %w", err)
	}

	return string(data), nil
}

// WriteRaw writes raw content to config.yaml
func WriteRaw(content string) error {
	// Validate YAML syntax first
	var test map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &test); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".claribot", "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}
