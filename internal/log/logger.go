package log

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

func NewLogger(level string, noColor bool) *slog.Logger {
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
