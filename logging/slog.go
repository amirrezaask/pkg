package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amirrezaask/pkg/env"
	"github.com/getsentry/sentry-go"
	slogmulti "github.com/samber/slog-multi"
	slogsentry "github.com/samber/slog-sentry/v2"
)

const (
	OutputFormat_JSON = "json"
	OutputFormat_Text = "text"
)

type Config struct {
	// Can be either a file name or stdout, empty also means stdout
	Output string
	// can be either 'json' or 'text'
	OutputFormat                      string
	DebugMode                         bool
	LogLevel                          slog.Level
	ChangeDebugModeWithSignalListener bool
	SentryConfig                      sentry.ClientOptions
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

// NewConfigFromEnv creates `Config` from env keys <LOG_OUTPUT_FORMAT|LOG_OUTPUT|LOG_LEVEL|LOG_SIGNAL_HANDLER|[SENTRY_DSN|SENTRY_ENV]>
// default values: LOG_OUTPUT_FORMAT: "json" | LOG_OUTPUT: "stdout" | LOG_LEVEL: "warn" | LOG_SIGNAL_HANDLER: "1".
func NewConfigFromEnv(appendSentry bool) Config {
	cfg := Config{
		Output:                            env.Default("LOG_OUTPUT", "stdout"),
		OutputFormat:                      env.Default("LOG_OUTPUT_FORMAT", OutputFormat_JSON),
		LogLevel:                          ParseLevel(env.Default("LOG_LEVEL", "warn")),
		ChangeDebugModeWithSignalListener: env.Default("LOG_SIGNAL_HANDLER", "1") == "1",
	}
	if appendSentry {
		cfg.SentryConfig = sentry.ClientOptions{Dsn: env.Default("SENTRY_DSN", ""), Environment: env.Default("SENTRY_ENV", "")}
	}
	return cfg
}

func Init(c Config) {
	var writer io.Writer
	if c.Output == "stdout" || c.Output == "" {
		writer = os.Stdout
	} else {
		file, err := os.OpenFile(c.Output, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}

		writer = file
	}

	handlers := []slog.Handler{}

	logLeveler := new(slog.LevelVar)
	logLeveler.Set(c.LogLevel)
	if c.OutputFormat == OutputFormat_JSON || c.OutputFormat == "" {
		handlers = append(handlers, slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level:     logLeveler,
			AddSource: true,
		}))
	} else if c.OutputFormat == OutputFormat_Text {
		handlers = append(handlers, slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level:     logLeveler,
			AddSource: true,
		}))
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
	if c.ChangeDebugModeWithSignalListener {
		if logLeveler.Level() == slog.LevelInfo || logLeveler.Level() == slog.LevelDebug {
			go changeDebugModeOnSignal(logLeveler)
		} else {
			slog.Warn("to use logging `ChangeDebugModeWithSignalListener` option, your initial log level must be `info` or `debug`")
		}
	}
}

func changeDebugModeOnSignal(levelVar *slog.LevelVar) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	debugMode := levelVar.Level() == slog.LevelDebug
	if debugMode {
		slog.Debug("debug mode is activated. Run `chll` (change log level) command in terminal to turn it off")
	}
	for {
		select {
		case <-sigs:
			if debugMode {
				levelVar.Set(slog.LevelInfo)
			} else {
				levelVar.Set(slog.LevelDebug)
			}
			debugMode = !debugMode
			slog.Info("updated debug mode", "active", debugMode)
		case <-context.Background().Done():
			slog.Info("stop listening for signals")
			return
		}
	}
}
