package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

type Logger struct {
	log *slog.Logger
}

func (l *Logger) DebugContext(ctx context.Context, format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	l.log.DebugContext(ctx, s)
}

func (l *Logger) InfoContext(ctx context.Context, format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	l.log.InfoContext(ctx, s)
}

func (l *Logger) WarnContext(ctx context.Context, format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	l.log.WarnContext(ctx, s)
}

func (l *Logger) ErrorContext(ctx context.Context, format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	l.log.ErrorContext(ctx, s)
}

func NewLogger(level string, noColor bool) *Logger {
	l := parseLevel(level)
	return &Logger{
		log: slog.New(tint.NewHandler(os.Stderr, &tint.Options{
			Level:     l,
			NoColor:   noColor,
			AddSource: l == slog.LevelDebug,
		})),
	}
}
