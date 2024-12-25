package destroy

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
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagForce = "force"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "destroy",
		ArgsUsage: "[repository]",
		Usage:     "Destroy an empty repository",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    FlagForce,
				Usage:   "force delete a non-empty repository",
				Aliases: []string{"f"},
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
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		err := fmt.Errorf("cannot create a JetStream context: %v", err)
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	client := repository.New(js)
	return destroyAction(logger, client)(ctx, c)
}

func destroyAction(logger *slog.Logger, client *repository.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		name := c.Args().Get(0)
		force := c.Bool(FlagForce)

		files, err := client.ListFiles(ctx, name)
		if err != nil {
			err := fmt.Errorf("cannot destroy repository '%s': %v", name, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		if len(files) > 0 && !force {
			err := fmt.Errorf("cannot destroy repository '%s': repository is not empty", name)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = client.Delete(ctx, name)
		if err != nil {
			err := fmt.Errorf("cannot destroy repository '%s': %v", name, err)
			logger.ErrorContext(ctx, err.Error())
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
