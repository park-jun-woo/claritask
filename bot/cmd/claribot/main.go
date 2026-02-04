package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"parkjunwoo.com/claribot/internal/handler"
	"parkjunwoo.com/claribot/pkg/claude"
	"parkjunwoo.com/claribot/pkg/telegram"

	"gopkg.in/yaml.v3"
)

const Version = "0.2.3"

// Router for command handling
var router *handler.Router

// Config represents the full configuration
type Config struct {
	Service struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"service"`
	Telegram struct {
		Token        string  `yaml:"token"`
		AllowedUsers []int64 `yaml:"allowed_users"`
	} `yaml:"telegram"`
	Claude struct {
		Timeout int `yaml:"timeout"` // seconds
		Max     int `yaml:"max"`
	} `yaml:"claude"`
	Project struct {
		Path string `yaml:"path"` // 프로젝트 생성 기본 경로
	} `yaml:"project"`
}

var bot *telegram.Bot

func main() {
	cfg := loadConfig()

	// Initialize Claude manager
	claude.Init(claude.Config{
		Timeout: time.Duration(cfg.Claude.Timeout) * time.Second,
		Max:     cfg.Claude.Max,
	})

	// Initialize router
	router = handler.NewRouter()

	// Log project path
	if cfg.Project.Path != "" {
		log.Printf("Project path: %s", cfg.Project.Path)
	}

	// Initialize Telegram bot
	if cfg.Telegram.Token != "" && cfg.Telegram.Token != "BOT_TOKEN" {
		var err error
		bot, err = telegram.New(cfg.Telegram.Token)
		if err != nil {
			log.Fatalf("Failed to create telegram bot: %v", err)
		}

		bot.SetHandler(handleTelegramMessage)
		bot.SetCallbackHandler(handleTelegramCallback)

		// Set menu commands
		bot.SetCommands([]telegram.Command{
			{Command: "start", Description: "시작"},
			{Command: "project", Description: "프로젝트 목록"},
		})

		if err := bot.Start(); err != nil {
			log.Fatalf("Failed to start telegram bot: %v", err)
		}
		log.Printf("Telegram bot connected: @%s", bot.Username())
	} else {
		log.Println("Telegram bot disabled (no token configured)")
	}

	// Setup HTTP handler
	http.HandleFunc("/", handleRequest)

	// Start HTTP server in goroutine
	addr := fmt.Sprintf("%s:%d", cfg.Service.Host, cfg.Service.Port)
	go func() {
		log.Printf("claribot v%s starting on %s", Version, addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down...")
	if bot != nil {
		bot.Stop()
	}
	log.Println("Goodbye!")
}

func loadConfig() Config {
	cfg := Config{}
	// defaults
	cfg.Service.Host = "127.0.0.1"
	cfg.Service.Port = 9847
	cfg.Claude.Timeout = 1200
	cfg.Claude.Max = 3

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: cannot get home directory: %v", err)
		return cfg
	}

	configPath := filepath.Join(homeDir, ".claribot", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Warning: cannot read config file: %v (using defaults)", err)
		return cfg
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("Warning: cannot parse config file: %v (using defaults)", err)
		return cfg
	}

	log.Printf("Config loaded from %s", configPath)
	return cfg
}

func handleTelegramMessage(msg telegram.Message) {
	log.Printf("[Telegram] %s: %s", msg.Username, msg.Text)

	switch msg.Text {
	case "/start":
		// Set reply keyboard
		bot.SetKeyboard(msg.ChatID, "Claribot 시작!", [][]string{
			{"!project list", "!status"},
		})

	case "/project":
		// Send project list with inline buttons
		bot.SendWithButtons(msg.ChatID, "프로젝트 선택:", [][]telegram.Button{
			{{Text: "claribot", Data: "switch:claribot"}},
		})

	default:
		// Handle ! commands via router
		if strings.HasPrefix(msg.Text, "!") {
			cmd := strings.TrimPrefix(msg.Text, "!")
			result := router.Execute(cmd)
			bot.Send(msg.ChatID, result.Message)
			return
		}

		// Handle message with current project context
		projectID, _ := router.GetProject()
		if projectID == "" {
			bot.Send(msg.ChatID, "프로젝트를 먼저 선택하세요.\n!project switch <id>")
			return
		}
		// Process message for current project
		bot.Send(msg.ChatID, fmt.Sprintf("[%s] %s", projectID, msg.Text))
	}
}

func handleTelegramCallback(cb telegram.Callback) {
	log.Printf("[Callback] %s: %s", cb.Username, cb.Data)

	// Handle project switch
	if strings.HasPrefix(cb.Data, "switch:") {
		projectID := strings.TrimPrefix(cb.Data, "switch:")
		result := router.Execute("project switch " + projectID)
		bot.AnswerCallback(cb.ID, projectID+" 선택됨")
		bot.Send(cb.ChatID, result.Message)
		return
	}

	bot.AnswerCallback(cb.ID, "")
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query().Get("args")

	if args == "" {
		http.Error(w, "missing args parameter", http.StatusBadRequest)
		return
	}

	log.Printf("Received: %s", args)

	// Echo back for now
	fmt.Fprint(w, args)
}
