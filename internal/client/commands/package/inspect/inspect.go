package inspect

import (
	"context"
	"fmt"
	"github.com/foohq/foojank/internal/client/actions"
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
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)

	return buildAction(logger)(ctx, c)
}

func buildAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		logger.Error("this feature is not yet implemented")
		return nil
	}
}
