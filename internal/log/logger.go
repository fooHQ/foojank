package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

var DefaultLogger = New("info", false)

func Debug(ctx context.Context, format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	DefaultLogger.DebugContext(ctx, s)
}

func Info(ctx context.Context, format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	DefaultLogger.InfoContext(ctx, s)
}

func Error(ctx context.Context, format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	DefaultLogger.ErrorContext(ctx, s)
}

func New(level string, noColor bool) *slog.Logger {
	l := parseLevel(level)
	return slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:     l,
		NoColor:   noColor,
		AddSource: l == slog.LevelDebug,
	}))
}

func parseLevel(level string) slog.Level {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
