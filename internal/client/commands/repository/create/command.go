package create

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/log"
	"github.com/foohq/foojank/internal/config"
)

const (
	FlagDescription = "description"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		ArgsUsage: "<name>",
		Usage:     "Create a new repository",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  FlagDescription,
				Usage: "set description",
			},
		},
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c, validateConfiguration)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)

	nc, err := server.New(logger, conf.Servers, conf.User.JWT, conf.User.KeySeed)
	if err != nil {
		err := fmt.Errorf("cannot connect to the server: %v", err)
		logger.Error(err.Error())
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
		description := c.String(FlagDescription)

		err := client.Create(ctx, name, description)
		if err != nil {
			err := fmt.Errorf("cannot create repository '%s': %v", name, err)
			logger.Error(err.Error())
			return err
		}

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.Servers == nil {
		return fmt.Errorf("servers not configured")
	}

	if conf.User == nil {
		return fmt.Errorf("user not configured")
	}

	return nil
}
