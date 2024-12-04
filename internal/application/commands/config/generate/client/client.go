package client

import (
	"context"
	"fmt"
	"github.com/foohq/foojank/internal/application/actions"
	"github.com/foohq/foojank/internal/application/commands/config/generate/seed"
	"github.com/urfave/cli/v3"
	"log/slog"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "client",
		ArgsUsage: "<seed-file>",
		Usage:     "Generate client configuration",
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

	return generateAction(logger)(ctx, c)
}

func generateAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		seedFile, err := seed.ParseOutput(c.Args().First())
		if err != nil {
			err = fmt.Errorf("cannot parse seed file: %v", err)
			logger.Error(err.Error())
			return err
		}

		clientFile, err := NewOutput(seedFile, "userTODO")
		if err != nil {
			err = fmt.Errorf("cannot generate client configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		fmt.Println(clientFile.String())

		return nil
	}
}
