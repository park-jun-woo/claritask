package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents log level
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// Logger is the global logger instance
type Logger struct {
	mu       sync.Mutex
	level    Level
	file     *os.File
	logger   *log.Logger
	filePath string
}

var (
	global     *Logger
	globalOnce sync.Once
)

// Config for logger initialization
type Config struct {
	Level    string // debug, info, warn, error
	FilePath string // empty = stdout only
}

// Init initializes the global logger
func Init(cfg Config) error {
	var initErr error
	globalOnce.Do(func() {
		global = &Logger{
			level: parseLevel(cfg.Level),
		}

		var writers []io.Writer
		writers = append(writers, os.Stdout)

		if cfg.FilePath != "" {
			// Ensure directory exists
			dir := filepath.Dir(cfg.FilePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				initErr = fmt.Errorf("create log directory: %w", err)
				return
			}

			file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				initErr = fmt.Errorf("open log file: %w", err)
				return
			}
			global.file = file
			global.filePath = cfg.FilePath
			writers = append(writers, file)
		}

		multiWriter := io.MultiWriter(writers...)
		global.logger = log.New(multiWriter, "", 0)

		// Redirect standard log package to the same output
		// so log.Printf() in all packages also goes to the log file
		log.SetOutput(multiWriter)
		log.SetFlags(0)
		log.SetPrefix("")
	})
	return initErr
}

// Close closes the log file
func Close() error {
	if global != nil && global.file != nil {
		return global.file.Close()
	}
	return nil
}

// GetLogger returns the global logger
func GetLogger() *Logger {
	if global == nil {
		Init(Config{Level: "info"})
	}
	return global
}

func parseLevel(s string) Level {
	switch s {
	case "debug":
		return DEBUG
	case "info", "":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := levelNames[level]
	msg := fmt.Sprintf(format, args...)
	l.logger.Printf("%s [%s] %s", timestamp, levelStr, msg)
}

// Debug logs debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Package-level convenience functions

func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}
