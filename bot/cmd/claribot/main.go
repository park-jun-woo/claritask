package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"parkjunwoo.com/claribot/internal/auth"
	"parkjunwoo.com/claribot/internal/config"
	"parkjunwoo.com/claribot/internal/db"
	"parkjunwoo.com/claribot/internal/handler"
	"parkjunwoo.com/claribot/internal/message"
	"parkjunwoo.com/claribot/internal/project"
	"parkjunwoo.com/claribot/internal/schedule"
	"parkjunwoo.com/claribot/internal/task"
	"parkjunwoo.com/claribot/internal/tghandler"
	"parkjunwoo.com/claribot/internal/types"
	"parkjunwoo.com/claribot/internal/webui"
	"parkjunwoo.com/claribot/pkg/claude"
	"parkjunwoo.com/claribot/pkg/logger"
	"parkjunwoo.com/claribot/pkg/telegram"
)

const Version = "0.2.21"

// Router for command handling
var router *handler.Router
var bot *telegram.Bot
var authService *auth.Auth
var bridgeManager *claude.BridgeManager
var startTime = time.Now()

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
	logger.Info("Global DB initialized")

	// Initialize auth service (keeps globalDB open for JWT/TOTP operations)
	authService = auth.New(globalDB)
	logger.Info("Auth service initialized")

	// Recover stuck messages from previous run
	if recovered, err := message.RecoverStuckMessages(1 * time.Hour); err != nil {
		logger.Error("Message recovery failed: %v", err)
	} else if recovered > 0 {
		logger.Info("Recovered %d stuck messages", recovered)
	}

	// Initialize Claude manager
	claude.Init(claude.Config{
		Timeout:    time.Duration(cfg.Claude.Timeout) * time.Second,
		MaxTimeout: time.Duration(cfg.Claude.MaxTimeout) * time.Second,
		Max:        cfg.Claude.Max,
	})
	logger.Info("Claude manager initialized (max=%d, timeout=%ds, max_timeout=%ds)", cfg.Claude.Max, cfg.Claude.Timeout, cfg.Claude.MaxTimeout)

	// Initialize Bridge manager (Agent SDK)
	if cfg.Bridge.Enabled {
		bridgeManager = claude.NewBridgeManager(claude.BridgeConfig{
			BridgePath:     cfg.Bridge.Path,
			NodePath:       cfg.Bridge.NodePath,
			IdleTimeout:    time.Duration(cfg.Bridge.IdleTimeout) * time.Second,
			PermissionMode: cfg.Bridge.PermissionMode,
		})
		logger.Info("Bridge manager initialized (path=%s, mode=%s)", cfg.Bridge.Path, cfg.Bridge.PermissionMode)
	} else {
		logger.Info("Bridge disabled (enable in config: bridge.enabled: true)")
	}

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

	// Restore last selected project
	router.RestoreProject()
	if id, _ := router.GetProject(); id != "" {
		logger.Info("Restored project: %s", id)
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
		tgHandler := tghandler.New(bot, router, cfg.Telegram.AllowedUsers)
		if bridgeManager != nil {
			tgHandler.SetBridgeManager(bridgeManager)
			logger.Info("Bridge integration enabled for Telegram")
		}
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

	// Initialize task notifier (reuse same notifier callback)
	task.Init(notifier)

	// Setup HTTP mux
	mux := http.NewServeMux()

	// Auth endpoints (bypass authMiddleware by path check)
	mux.HandleFunc("/api/auth/setup", handleAuthSetup)
	mux.HandleFunc("/api/auth/login", handleAuthLogin)
	mux.HandleFunc("/api/auth/status", handleAuthStatus)
	mux.HandleFunc("/api/auth/logout", handleAuthLogout)
	mux.HandleFunc("/api/auth/totp-setup", handleAuthTOTPSetup)

	// RESTful API endpoints
	router.RegisterRESTfulRoutes(mux)

	// Legacy API endpoints (backward compatibility for CLI/Telegram)
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api", handleAPIRequest)

	// Web UI: all non-API GET requests serve static files
	webuiHandler := webui.Handler()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// GET with ?args= = CLI backward compatibility
		if r.Method == http.MethodGet && r.URL.Query().Get("args") != "" {
			handleAPIRequest(w, r)
			return
		}

		// Everything else: serve Web UI
		webuiHandler.ServeHTTP(w, r)
	})

	// Middleware chain: CORS → Auth → mux
	handler := corsMiddleware(authMiddleware(mux))

	// Start HTTP server in goroutine with timeout settings
	addr := fmt.Sprintf("%s:%d", cfg.Service.Host, cfg.Service.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Minute, // Long timeout for Claude operations
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		logger.Info("HTTP server starting on %s", addr)
		logger.Info("Web UI available at http://%s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	// Shutdown Bridge processes
	if bridgeManager != nil {
		activeBridges := bridgeManager.ActiveBridges()
		if activeBridges > 0 {
			logger.Info("Closing %d active Bridge processes...", activeBridges)
			bridgeManager.Shutdown()
			logger.Info("Bridge processes closed")
		}
	}

	// Shutdown Claude sessions
	activeSessions := claude.ActiveSessions()
	if activeSessions > 0 {
		logger.Info("Closing %d active Claude sessions...", activeSessions)
		claude.Shutdown()
		logger.Info("Claude sessions closed")
	}

	globalDB.Close()
	logger.Info("Goodbye!")
}

// handleAPIRequest handles command requests from CLI and Web UI
func handleAPIRequest(w http.ResponseWriter, r *http.Request) {
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
		logger.Debug("[API/POST] %s", cmdStr)
	} else if r.Method == http.MethodGet {
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
		logger.Debug("[API/GET] %s", cmdStr)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(types.Result{
			Success: false,
			Message: "method not allowed",
		})
		return
	}

	snapshot := router.SnapshotContext()
	result := router.Execute(snapshot, cmdStr)

	if !result.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(result)
}

// handleHealth returns service health information
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uptime := time.Since(startTime).Seconds()
	claudeStatus := claude.GetStatus()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"version": Version,
		"uptime":  int(uptime),
		"claude": map[string]interface{}{
			"used":      claudeStatus.Used,
			"max":       claudeStatus.Max,
			"available": claudeStatus.Available,
		},
	})
}

// corsMiddleware adds CORS headers for development (Vite dev server)
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && strings.HasPrefix(origin, "http://localhost") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authMiddleware checks authentication for non-local requests.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Localhost bypass
		remoteIP := r.RemoteAddr
		if host, _, err := net.SplitHostPort(remoteIP); err == nil {
			remoteIP = host
		}
		if remoteIP == "127.0.0.1" || remoteIP == "::1" {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Auth endpoints pass through
		if strings.HasPrefix(r.URL.Path, "/api/auth/") {
			next.ServeHTTP(w, r)
			return
		}

		// 3. Allow non-API requests (Web UI static files) to pass through
		if !strings.HasPrefix(r.URL.Path, "/api") {
			next.ServeHTTP(w, r)
			return
		}

		// 4. If setup not completed, block API requests
		if !authService.IsSetupCompleted() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "setup not completed"})
			return
		}

		// 5. Validate JWT cookie for API requests
		if !authService.IsAuthenticated(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "authentication required"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleAuthSetup handles POST /api/auth/setup
func handleAuthSetup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		Password string `json:"password"`
		TOTPCode string `json:"totp_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}

	result, err := authService.Setup(req.Password, req.TOTPCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := map[string]interface{}{}
	if result.TOTPURI != "" {
		resp["totp_uri"] = result.TOTPURI
	}
	if result.Token != "" {
		auth.SetTokenCookie(w, result.Token)
		resp["success"] = true
		resp["token"] = result.Token
		resp["expires_at"] = time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	}
	json.NewEncoder(w).Encode(resp)
}

// handleAuthLogin handles POST /api/auth/login
func handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		Password string `json:"password"`
		TOTPCode string `json:"totp_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}

	token, err := authService.Login(req.Password, req.TOTPCode)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	auth.SetTokenCookie(w, token)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"token":      token,
		"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
}

// handleAuthStatus handles GET /api/auth/status
func handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authService.Status(r))
}

// handleAuthLogout handles POST /api/auth/logout
func handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	auth.ClearTokenCookie(w)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleAuthTOTPSetup handles GET /api/auth/totp-setup
func handleAuthTOTPSetup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	uri, err := authService.GetTOTPSetupURI()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"totp_uri": uri})
}
