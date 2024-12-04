package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/config/generate/seed"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "agent",
		ArgsUsage: "<seed-file>",
		Usage:     "Generate agent configuration",
		Action:    action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)

	return generateAction(logger)(ctx, c)
}

func generateAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		seedFile, err := seed.ParseOutput(c.Args().First())
		if err != nil {
			err = fmt.Errorf("cannot parse seed file: %v", err)
			logger.Error(err.Error())
			return err
		}

		username := fmt.Sprintf("AG%s", nuid.Next())
		clientFile, err := NewOutput(seedFile, username)
		if err != nil {
			err = fmt.Errorf("cannot generate client configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		fmt.Println(clientFile.String())

		return nil
	}
}
