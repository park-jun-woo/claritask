package bot

import (
	"time"

	"claritask/bot/internal/config"
	"claritask/bot/internal/service"

	"github.com/rs/zerolog"
	"gopkg.in/telebot.v3"
)

// Bot represents the telegram bot
type Bot struct {
	tg      *telebot.Bot
	cfg     *config.Config
	svc     *service.Service
	state   *StateManager
	limiter *RateLimiter
	logger  zerolog.Logger
}

// New creates a new bot instance
func New(cfg *config.Config, logger zerolog.Logger) (*Bot, error) {
	// Create telegram bot
	pref := telebot.Settings{
		Token:  cfg.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	tg, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	// Create service
	svc, err := service.New(cfg.GetDBPath())
	if err != nil {
		return nil, err
	}

	// Create bot
	b := &Bot{
		tg:      tg,
		cfg:     cfg,
		svc:     svc,
		state:   NewStateManager(30 * time.Minute),
		limiter: NewRateLimiter(cfg.RateLimit, cfg.RateBurst),
		logger:  logger,
	}

	// Setup routes
	b.SetupRoutes()

	return b, nil
}

// Start starts the bot (blocking)
func (b *Bot) Start() {
	b.logger.Info().Msg("bot started")
	b.tg.Start()
}

// Stop stops the bot gracefully
func (b *Bot) Stop() {
	b.logger.Info().Msg("stopping bot...")
	b.tg.Stop()
	if b.svc != nil {
		b.svc.Close()
	}
}

// getCurrentProject gets the current project for a user
func (b *Bot) getCurrentProject(userID int64) string {
	state := b.state.Get(userID)
	if state.CurrentProject != "" {
		return state.CurrentProject
	}

	// Fallback to DB current project
	project, err := b.svc.GetCurrentProject()
	if err != nil || project == nil {
		return ""
	}

	// Cache it
	b.state.SetCurrentProject(userID, project.ID)
	return project.ID
}
