package create

import (
	"context"
	"fmt"
	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v3"
	"log/slog"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		ArgsUsage: "<name>",
		Usage:     "Create a new repository",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "description",
			},
		},
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)

	nc, err := actions.NewServerConnection(ctx, conf, logger)
	if err != nil {
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		err := fmt.Errorf("cannot create a JetStream context: %v", err)
		logger.Error(err.Error())
		return err
	}

	client := repository.New(js)
	return createAction(logger, client)(ctx, c)
}

func createAction(logger *slog.Logger, client *repository.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		name := c.Args().Get(0)
		description := c.String("description")

		err := client.Create(ctx, name, description)
		if err != nil {
			err := fmt.Errorf("cannot create repository '%s': %v", name, err)
			logger.Error(err.Error())
			return err
		}

		return nil
	}
}
