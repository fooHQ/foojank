package server

import (
	"context"
	"fmt"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/config/generate/seed"
	"github.com/urfave/cli/v3"
	"log/slog"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "server",
		ArgsUsage: "<seed-file>",
		Usage:     "Generate server configuration",
		Action:    action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)

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

		serverFile, err := NewOutput(seedFile)
		if err != nil {
			err = fmt.Errorf("cannot generate server configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		fmt.Println(serverFile.String())

		return nil
	}
}
