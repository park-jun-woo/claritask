package bot

import (
	"gopkg.in/telebot.v3"
)

// SetupRoutes configures all bot routes
func (b *Bot) SetupRoutes() {
	// Global middleware
	b.tg.Use(b.AuditMiddleware())
	b.tg.Use(b.RateLimitMiddleware())
	b.tg.Use(b.AuthMiddleware())

	// Basic commands
	b.tg.Handle("/start", b.HandleStart)
	b.tg.Handle("/help", b.HandleHelp)
	b.tg.Handle("/status", b.HandleStatus)

	// Project commands
	b.tg.Handle("/project", b.HandleProject)

	// Task commands
	b.tg.Handle("/task", b.HandleTask)

	// Message commands
	b.tg.Handle("/msg", b.HandleMessage)

	// Expert commands
	b.tg.Handle("/expert", b.HandleExpert)

	// Settings commands
	b.tg.Handle("/settings", b.HandleSettings)

	// Callback queries (inline buttons)
	b.tg.Handle(telebot.OnCallback, b.HandleCallback)

	// Text messages (conversational input)
	b.tg.Handle(telebot.OnText, b.HandleText)
}
