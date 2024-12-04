package actions

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/server/flags"
)

func CommandNotFound(ctx context.Context, c *cli.Command, s string) {
	logger := NewLogger(ctx, c)
	msg := fmt.Sprintf("command '%s %s' does not exist", c.FullName(), s)
	logger.Error(msg)
	os.Exit(1)
}

func NewLogger(ctx context.Context, c *cli.Command) *slog.Logger {
	level := c.Int(flags.LogLevel)
	noColor := c.Bool(flags.NoColor)
	return slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:     slog.Level(level),
		NoColor:   noColor,
		AddSource: true,
	}))
}
