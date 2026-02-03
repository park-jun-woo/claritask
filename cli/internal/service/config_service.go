package service

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TTYConfig holds TTY-related configuration
type TTYConfig struct {
	MaxParallelSessions int `yaml:"max_parallel_sessions"`
	TerminalCloseDelay  int `yaml:"terminal_close_delay"`
}

// VSCodeConfig holds VSCode extension configuration
type VSCodeConfig struct {
	SyncInterval      int  `yaml:"sync_interval"`
	WatchFeatureFiles bool `yaml:"watch_feature_files"`
}

// Config holds all Claritask configuration
type Config struct {
	TTY    TTYConfig    `yaml:"tty"`
	VSCode VSCodeConfig `yaml:"vscode"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		TTY: TTYConfig{
			MaxParallelSessions: 3,
			TerminalCloseDelay:  1,
		},
		VSCode: VSCodeConfig{
			SyncInterval:      1000,
			WatchFeatureFiles: true,
		},
	}
}

// LoadConfig loads configuration from .claritask/config.yaml
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Try to load from file
	configPath := filepath.Join(".claritask", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// File doesn't exist, return defaults
		return config, nil
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return config, nil // Return defaults on parse error
	}

	// Validate and clamp values
	if config.TTY.MaxParallelSessions < 1 {
		config.TTY.MaxParallelSessions = 1
	}
	if config.TTY.MaxParallelSessions > 10 {
		config.TTY.MaxParallelSessions = 10
	}

	if config.VSCode.SyncInterval < 100 {
		config.VSCode.SyncInterval = 100
	}

	return config, nil
}

// LoadConfigFrom loads configuration from a specific directory
func LoadConfigFrom(dir string) (*Config, error) {
	config := DefaultConfig()

	configPath := filepath.Join(dir, ".claritask", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, nil
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return config, nil
	}

	// Validate
	if config.TTY.MaxParallelSessions < 1 {
		config.TTY.MaxParallelSessions = 1
	}
	if config.TTY.MaxParallelSessions > 10 {
		config.TTY.MaxParallelSessions = 10
	}

	return config, nil
}
