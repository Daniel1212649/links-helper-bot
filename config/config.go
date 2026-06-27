package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"time"
)

const (
	defaultTelegramAPIHost = "api.telegram.org"
	defaultPollBatchSize   = 100
	defaultPollInterval    = time.Second
	defaultHTTPTimeout     = 35 * time.Second
)

type Config struct {
	TelegramBotToken string
	TelegramAPIHost  string
	DatabaseURL      string
	AppEnv           string
	LogLevel         string
	PollBatchSize    int
	PollInterval     time.Duration
	HTTPTimeout      time.Duration
}

func Load() (Config, error) {
	tokenFlag := flag.String("tg-bot-token", "", "Telegram bot token")
	flag.Parse()

	cfg := Config{
		TelegramBotToken: env("TELEGRAM_BOT_TOKEN", ""),
		TelegramAPIHost:  env("TELEGRAM_API_HOST", defaultTelegramAPIHost),
		DatabaseURL:      env("DATABASE_URL", ""),
		AppEnv:           env("APP_ENV", "local"),
		LogLevel:         env("LOG_LEVEL", "info"),
		PollBatchSize:    envInt("POLL_BATCH_SIZE", defaultPollBatchSize),
		PollInterval:     envDuration("POLL_INTERVAL", defaultPollInterval),
		HTTPTimeout:      envDuration("HTTP_TIMEOUT", defaultHTTPTimeout),
	}

	if *tokenFlag != "" {
		cfg.TelegramBotToken = *tokenFlag
	}

	if cfg.TelegramBotToken == "" {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.PollBatchSize <= 0 {
		return Config{}, errors.New("POLL_BATCH_SIZE must be positive")
	}

	return cfg, nil
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err == nil {
		return parsed
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
