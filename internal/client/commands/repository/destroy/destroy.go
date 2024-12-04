package destroy

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
		Name:      "destroy",
		ArgsUsage: "[repository]",
		Usage:     "Destroy an empty repository",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Usage:   "force delete of non-empty repository",
				Aliases: []string{"f"},
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
	return destroyAction(logger, client)(ctx, c)
}

func destroyAction(logger *slog.Logger, client *repository.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		name := c.Args().Get(0)
		force := c.Bool("force")

		files, err := client.ListFiles(ctx, name)
		if err != nil {
			err := fmt.Errorf("cannot destroy repository '%s': %v", name, err)
			logger.Error(err.Error())
			return err
		}

		if len(files) > 0 && !force {
			err := fmt.Errorf("cannot destroy repository '%s': repository is not empty", name)
			logger.Error(err.Error())
			return err
		}

		err = client.Delete(ctx, name)
		if err != nil {
			err := fmt.Errorf("cannot destroy repository '%s': %v", name, err)
			logger.Error(err.Error())
			return err
		}

		return nil
	}
}
