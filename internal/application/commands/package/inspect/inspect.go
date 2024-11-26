package inspect

import (
	"context"
	"fmt"
	"github.com/foohq/foojank/internal/application/actions"
	"github.com/urfave/cli/v3"
	"log/slog"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "inspect",
		ArgsUsage: "<package>",
		Usage:     "Examine the contents of a package",
		Action:    action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	logger := actions.NewLogger(ctx, c)

	if c.Args().Len() != 1 {
		err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
		logger.Error(err.Error())
		return err
	}

	return buildAction(logger)(ctx, c)
}

func buildAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		logger.Error("this feature is not yet implemented")
		return nil
	}
}
