package logging

import (
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
	slogmulti "github.com/samber/slog-multi"
	slogsentry "github.com/samber/slog-sentry/v2"
)

type Config struct {
	DebugMode    bool
	LogLevel     slog.Level
	SentryConfig sentry.ClientOptions
}

func ParseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelError
	}
}

func Init(c Config) {
	handlers := []slog.Handler{
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     c.LogLevel,
			AddSource: true,
		}),
	}

	if c.SentryConfig.Dsn != "" && c.SentryConfig.Environment != "" {
		err := sentry.Init(c.SentryConfig)
		if err != nil {
			panic(err)
		}
		handlers = append(handlers, slogsentry.Option{
			Level:     slog.LevelWarn,
			AddSource: true,
		}.NewSentryHandler())
	}

	logger := slog.New(
		slogmulti.Fanout(
			handlers...,
		),
	)

	slog.SetDefault(logger)
}
