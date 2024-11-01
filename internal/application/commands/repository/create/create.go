package create

import (
	"context"
	"fmt"
	"github.com/foojank/foojank/clients/repository"
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v3"
	"log/slog"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "create",
		Description: "Create a repository",
		ArgsUsage:   "<name>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "description",
			},
		},
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	logger := actions.NewLogger(ctx, c)

	if c.Args().Len() != 1 {
		err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
		logger.Error(err.Error())
		return err
	}

	nc, err := actions.NewNATSConnection(ctx, c, logger)
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
