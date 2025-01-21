package create

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagDescription = "description"
	FlagServer      = flags.Server
	FlagUserJWT     = flags.UserJWT
	FlagUserKey     = flags.UserKey
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
			&cli.StringSliceFlag{
				Name:    FlagServer,
				Usage:   "set server URL",
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:  FlagUserJWT,
				Usage: "set user JWT token",
			},
			&cli.StringFlag{
				Name:  FlagUserKey,
				Usage: "set user secret key",
			},
		},
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: cannot parse configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)

	nc, err := server.New(logger, conf.Client.Server, *conf.Client.UserJWT, *conf.Client.UserKey)
	if err != nil {
		err := fmt.Errorf("cannot connect to the server: %w", err)
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		err := fmt.Errorf("cannot create a JetStream context: %w", err)
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	client := repository.New(js)
	return createAction(logger, client)(ctx, c)
}

func createAction(logger *slog.Logger, client *repository.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		name := c.Args().Get(0)
		description := c.String(FlagDescription)

		err := client.Create(ctx, name, description)
		if err != nil {
			err := fmt.Errorf("cannot create repository '%s': %w", name, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.LogLevel == nil {
		return errors.New("log level not configured")
	}

	if conf.NoColor == nil {
		return errors.New("no color not configured")
	}

	if conf.Client == nil {
		return errors.New("client configuration is missing")
	}

	if len(conf.Client.Server) == 0 {
		return errors.New("server not configured")
	}

	if conf.Client.UserJWT == nil {
		return errors.New("user jwt not configured")
	}

	if conf.Client.UserKey == nil {
		return errors.New("user key not configured")
	}

	return nil
}
