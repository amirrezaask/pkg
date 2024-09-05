package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
)

type Config struct {
	DebugMode bool
	LogLevel  slog.Level
}

const (
	std_context_contextual_info_key = "___std_contextual_info___"
)

var (
	//You can set this variable inside your config code if you have another way of doing configuration.
	RuntimeFileInfo = false
)

type LogLevel = slog.Level

func Init(c Config) {
	globalLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     c.LogLevel,
	}))

	if !c.DebugMode {
		globalLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: false,
			Level:     c.LogLevel,
		}))
	}
	slog.SetDefault(globalLogger)
	slog.SetLogLoggerLevel(c.LogLevel)
}

func DebugContext(ctx context.Context, msg string, kvs ...any) {
	if RuntimeFileInfo {
		pc, file, line, ok := runtime.Caller(1)
		fName := runtime.FuncForPC(pc)
		if ok {
			kvs = append(kvs, "function", fName)
			kvs = append(kvs, "file", file)
			kvs = append(kvs, "line", line)
		}
	}

	if ctxInfo := ctx.Value(std_context_contextual_info_key); ctxInfo != nil {
		if ctxInfoMap, ok := ctxInfo.(map[string]any); ok {
			for k, v := range ctxInfoMap {
				kvs = append(kvs, k, v)
			}
		}

	}
	slog.Debug(msg, kvs...)
}

func Debug(msg string, kvs ...any) {
	if RuntimeFileInfo {
		pc, file, line, ok := runtime.Caller(1)
		fName := runtime.FuncForPC(pc)
		if ok {
			kvs = append(kvs, "function", fName)
			kvs = append(kvs, "file", file)
			kvs = append(kvs, "line", line)
		}
	}
	slog.Debug(msg, kvs...)
}

func Error(msg string, kvs ...any) {
	if RuntimeFileInfo {
		pc, file, line, ok := runtime.Caller(1)
		fName := runtime.FuncForPC(pc).Name()
		if ok {
			kvs = append(kvs, "function", fName)
			kvs = append(kvs, "file", file)
			kvs = append(kvs, "line", line)
		}
	}

	slog.Error(msg, kvs...)
}

func ErrorContext(ctx context.Context, msg string, kvs ...any) {
	if RuntimeFileInfo {
		pc, file, line, ok := runtime.Caller(1)
		fName := runtime.FuncForPC(pc).Name()
		if ok {
			kvs = append(kvs, "function", fName)
			kvs = append(kvs, "file", file)
			kvs = append(kvs, "line", line)
		}
	}

	if ctxInfo := ctx.Value(std_context_contextual_info_key); ctxInfo != nil {
		if ctxInfoMap, ok := ctxInfo.(map[string]any); ok {
			for k, v := range ctxInfoMap {
				kvs = append(kvs, k, v)
			}
		}

	}
	slog.Error(msg, kvs...)
}

func Warn(msg string, kvs ...any) {
	if RuntimeFileInfo {
		pc, file, line, ok := runtime.Caller(1)
		fName := runtime.FuncForPC(pc)
		if ok {
			kvs = append(kvs, "function", fName)
			kvs = append(kvs, "file", file)
			kvs = append(kvs, "line", line)
		}
	}

	slog.Warn(msg, kvs...)
}

func LogWhenError(err error, msgAndArgs ...any) error {
	if err != nil {
		if len(msgAndArgs) < 1 {
			msgAndArgs = append(msgAndArgs, "No message was provided")
		}
		msgAndArgs = append(msgAndArgs, "err", err)

		if _, isString := msgAndArgs[0].(string); !isString {
			msgAndArgs[0] = fmt.Sprint(msgAndArgs[0])
		}
		Error(msgAndArgs[0].(string), msgAndArgs[1:])
	}

	return err
}
func LogWhenError1[T any](t T, err error, msgAndArgs ...any) (T, error) {
	if err != nil {
		if len(msgAndArgs) < 1 {
			msgAndArgs = append(msgAndArgs, "No message was provided")
			msgAndArgs = append(msgAndArgs, "err", err)
		}
		if _, isString := msgAndArgs[0].(string); !isString {
			msgAndArgs[0] = fmt.Sprint(msgAndArgs[0])
		}
		Error(msgAndArgs[0].(string), msgAndArgs[1:])

	}

	return t, err
}

func WithContextualInfo(ctx context.Context, kvs ...any) context.Context {
	if len(kvs)%2 == 0 {
		return ctx
	}

	info := map[string]any{}

	for i := 0; i < len(kvs); i++ {
		if i%2 == 0 {
			info[fmt.Sprint(kvs[i])] = kvs[i+1]
		}
	}

	if ctx.Value(std_context_contextual_info_key) != nil {
		oldInfo := ctx.Value(std_context_contextual_info_key).(map[string]any)
		for k, v := range oldInfo {
			if _, exists := info[k]; !exists {
				info[k] = v
			}
		}
	}

	return context.WithValue(ctx, std_context_contextual_info_key, info)
}
