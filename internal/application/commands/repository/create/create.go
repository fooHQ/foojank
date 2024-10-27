package create

import (
	"fmt"
	"github.com/foojank/foojank/clients/repository"
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v2"
	"log/slog"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "create",
		Description: "Create a repository",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "description",
			},
		},
		Action: action,
	}
}

func action(c *cli.Context) error {
	logger := actions.NewLogger(c)
	nc, err := actions.NewNATSConnection(c, logger)
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
	return createAction(logger, client)(c)
}

func createAction(logger *slog.Logger, client *repository.Client) cli.ActionFunc {
	return func(c *cli.Context) error {
		name := c.Args().Get(0)
		description := c.String("description")

		if name == "" {
			return fmt.Errorf("command expects an argument")
		}

		err := client.Create(c.Context, name, description)
		if err != nil {
			return err
		}

		return nil
	}
}
