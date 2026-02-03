package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v9"
)

// Config holds all configuration for the bot
type Config struct {
	// Telegram settings
	TelegramToken string  `env:"TELEGRAM_TOKEN,required"`
	AllowedUsers  []int64 `env:"ALLOWED_USERS" envSeparator:","`
	AdminUsers    []int64 `env:"ADMIN_USERS" envSeparator:","`

	// Database settings
	DBPath string `env:"CLARITASK_DB" envDefault:"~/.claritask/db.clt"`

	// Notification settings
	NotifyOnComplete bool `env:"NOTIFY_ON_COMPLETE" envDefault:"true"`
	NotifyOnFail     bool `env:"NOTIFY_ON_FAIL" envDefault:"true"`

	// Logging
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	// Rate limiting
	RateLimit float64 `env:"RATE_LIMIT" envDefault:"1"`
	RateBurst int     `env:"RATE_BURST" envDefault:"5"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// IsAllowed checks if a user is allowed to use the bot
func (c *Config) IsAllowed(userID int64) bool {
	// Check admin users first
	if c.IsAdmin(userID) {
		return true
	}

	// Check allowed users
	for _, id := range c.AllowedUsers {
		if id == userID {
			return true
		}
	}
	return false
}

// IsAdmin checks if a user is an admin
func (c *Config) IsAdmin(userID int64) bool {
	for _, id := range c.AdminUsers {
		if id == userID {
			return true
		}
	}
	return false
}

// GetDBPath returns the expanded database path
func (c *Config) GetDBPath() string {
	path := c.DBPath

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}

	return path
}
