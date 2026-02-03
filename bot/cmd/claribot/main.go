package main

import (
	"os"
	"os/signal"
	"syscall"

	"claritask/bot/internal/bot"
	"claritask/bot/internal/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	// Setup logger
	setupLogger()

	log.Info().
		Str("version", Version).
		Str("commit", Commit).
		Str("build_time", BuildTime).
		Msg("starting claribot")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Create bot
	b, err := bot.New(cfg, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create bot")
	}

	// Handle shutdown signals
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("received shutdown signal")
		b.Stop()
	}()

	// Start bot (blocking)
	b.Start()

	log.Info().Msg("claribot stopped")
}

func setupLogger() {
	// Console writer for development
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05",
	}

	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Logger()

	// Set log level from environment
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
