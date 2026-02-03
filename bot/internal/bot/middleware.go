package bot

import (
	"time"

	"gopkg.in/telebot.v3"
)

// AuthMiddleware checks if user is allowed to use the bot
func (b *Bot) AuthMiddleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			userID := c.Sender().ID
			if !b.cfg.IsAllowed(userID) {
				b.logger.Warn().
					Int64("user_id", userID).
					Str("username", c.Sender().Username).
					Str("first_name", c.Sender().FirstName).
					Msg("unauthorized access attempt")
				return c.Send("❌ 권한이 없습니다.")
			}
			return next(c)
		}
	}
}

// AdminOnlyMiddleware checks if user is an admin
func (b *Bot) AdminOnlyMiddleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			userID := c.Sender().ID
			if !b.cfg.IsAdmin(userID) {
				return c.Send("❌ 관리자 권한이 필요합니다.")
			}
			return next(c)
		}
	}
}

// AuditMiddleware logs all commands
func (b *Bot) AuditMiddleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			start := time.Now()

			err := next(c)

			b.logger.Info().
				Int64("user_id", c.Sender().ID).
				Str("username", c.Sender().Username).
				Str("text", c.Text()).
				Dur("duration", time.Since(start)).
				Err(err).
				Msg("command")

			return err
		}
	}
}

// RateLimitMiddleware limits request rate per user
func (b *Bot) RateLimitMiddleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			if !b.limiter.Allow(c.Sender().ID) {
				return c.Send("⏳ 잠시 후 다시 시도해주세요.")
			}
			return next(c)
		}
	}
}
