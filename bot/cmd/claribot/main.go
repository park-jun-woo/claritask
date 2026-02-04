package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"parkjunwoo.com/claribot/internal/config"
	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/handler"
	"parkjunwoo.com/claribot/internal/project"
	"parkjunwoo.com/claribot/internal/schedule"
	"parkjunwoo.com/claribot/internal/tghandler"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/pkg/claude"
	"parkjunwoo.com/claribot/pkg/logger"
	"parkjunwoo.com/claribot/pkg/telegram"
)

const Version = "0.2.19"

// Router for command handling
var router *handler.Router
var bot *telegram.Bot

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: config load error: %v (using defaults)\n", err)
	}

	// Validate config
	warnings := cfg.Validate()
	for _, w := range warnings {
		fmt.Printf("Config warning: %s\n", w)
	}

	// Initialize logger
	if err := logger.Init(logger.Config{
		Level:    cfg.Log.Level,
		FilePath: cfg.GetLogFilePath(),
	}); err != nil {
		fmt.Printf("Warning: logger init error: %v\n", err)
	}
	defer logger.Close()

	logger.Info("claribot v%s starting...", Version)

	// Initialize global DB
	globalDB, err := db.OpenGlobal()
	if err != nil {
		logger.Error("Failed to open global DB: %v", err)
		os.Exit(1)
	}
	if err := globalDB.MigrateGlobal(); err != nil {
		logger.Error("Failed to migrate global DB: %v", err)
		os.Exit(1)
	}
	globalDB.Close()
	logger.Info("Global DB initialized")

	// Initialize Claude manager
	claude.Init(claude.Config{
		Timeout: time.Duration(cfg.Claude.Timeout) * time.Second,
		Max:     cfg.Claude.Max,
	})
	logger.Info("Claude manager initialized (max=%d, timeout=%ds)", cfg.Claude.Max, cfg.Claude.Timeout)

	// Initialize router
	router = handler.NewRouter()

	// Set pagination page size
	if cfg.Pagination.PageSize > 0 {
		router.SetPageSize(cfg.Pagination.PageSize)
		logger.Debug("Page size: %d", cfg.Pagination.PageSize)
	}

	// Set project default path
	if cfg.Project.Path != "" {
		project.SetDefaultPath(cfg.Project.Path)
		logger.Info("Project path: %s", project.DefaultPath)
	}

	// Initialize Telegram bot
	if cfg.IsTelegramEnabled() {
		var err error
		bot, err = telegram.New(cfg.Telegram.Token)
		if err != nil {
			logger.Error("Failed to create telegram bot: %v", err)
			os.Exit(1)
		}

		// Set admin chat ID from config (for schedule notifications)
		if cfg.Telegram.AdminChatID != 0 {
			bot.SetAdminChatID(cfg.Telegram.AdminChatID)
			logger.Info("Admin chat ID configured: %d", cfg.Telegram.AdminChatID)
		}

		// Setup handler (also registers menu commands)
		tgHandler := tghandler.New(bot, router)
		bot.SetHandler(tgHandler.HandleMessage)
		bot.SetCallbackHandler(tgHandler.HandleCallback)

		if err := bot.Start(); err != nil {
			logger.Error("Failed to start telegram bot: %v", err)
			os.Exit(1)
		}
		logger.Info("Telegram bot connected: @%s", bot.Username())
	} else {
		logger.Info("Telegram bot disabled (no token configured)")
	}

	// Initialize scheduler with telegram notifier
	notifier := func(projectID *string, msg string) {
		if bot != nil {
			if err := bot.Broadcast(msg); err != nil {
				logger.Error("Schedule notification failed: %v", err)
			}
		}
	}
	if err := schedule.Init(notifier); err != nil {
		logger.Error("Failed to initialize scheduler: %v", err)
	} else {
		logger.Info("Scheduler initialized (jobs: %d)", schedule.JobCount())
	}

	// Setup HTTP handler
	http.HandleFunc("/", handleRequest)

	// Start HTTP server in goroutine
	addr := fmt.Sprintf("%s:%d", cfg.Service.Host, cfg.Service.Port)
	go func() {
		logger.Info("HTTP server starting on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("HTTP server error: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Shutting down...")

	// Graceful shutdown
	schedule.Shutdown()
	logger.Info("Scheduler stopped")

	if bot != nil {
		bot.Stop()
		logger.Info("Telegram bot stopped")
	}

	// Shutdown Claude sessions
	activeSessions := claude.ActiveSessions()
	if activeSessions > 0 {
		logger.Info("Closing %d active Claude sessions...", activeSessions)
		claude.Shutdown()
		logger.Info("Claude sessions closed")
	}

	logger.Info("Goodbye!")
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var cmdStr string

	if r.Method == http.MethodPost {
		// POST: JSON body
		var req types.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(types.Result{
				Success: false,
				Message: "invalid JSON: " + err.Error(),
			})
			return
		}
		cmdStr = req.ToCommandString()
		logger.Debug("[CLI/POST] %s", cmdStr)
	} else {
		// GET: query parameter (backward compatibility)
		cmdStr = r.URL.Query().Get("args")
		if cmdStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(types.Result{
				Success: false,
				Message: "missing args parameter",
			})
			return
		}
		logger.Debug("[CLI/GET] %s", cmdStr)
	}

	result := router.Execute(cmdStr)

	if !result.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(result)
}
