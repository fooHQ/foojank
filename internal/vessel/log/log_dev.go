//go:build dev

package log

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

var logger = slog.New(tint.NewHandler(os.Stderr, &tint.Options{
	Level:     slog.LevelDebug,
	AddSource: true,
}))

var Debug = logger.Debug
