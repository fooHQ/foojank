package list

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
		Name:        "list",
		ArgsUsage:   "[repository]",
		Description: "List repositories or their contents.",
		Action:      action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	logger := actions.NewLogger(ctx, c)
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
	return listAction(logger, client)(ctx, c)
}

func listAction(logger *slog.Logger, client *repository.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() > 0 {
			for _, r := range c.Args().Slice() {
				files, err := client.ListFiles(ctx, r)
				if err != nil {
					err := fmt.Errorf("cannot list contents of repository '%s': %v", r, err)
					logger.Error(err.Error())
					return err
				}

				for _, file := range files {
					fmt.Printf("%#v\n", file)
				}
			}
			return nil
		}

		repos, err := client.List(ctx)
		if err != nil {
			return err
		}

		for i := range repos {
			fmt.Printf("%#v\n", repos[i])
		}

		return nil
	}
}
