package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config/v2"
	"github.com/foohq/foojank/internal/log"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "server",
		ArgsUsage: "<master-config>",
		Usage:     "Generate server config from master config",
		Action:    action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewClientConfig(ctx, c)
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
	return generateAction(logger)(ctx, c)
}

func generateAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		confFile := c.Args().First()
		confInput, err := config.ParseFile(confFile)
		if err != nil {
			err := fmt.Errorf("cannot parse master configuration file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = validateInputConfiguration(confInput)
		if err != nil {
			err := fmt.Errorf("invalid master configuration file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		var confServer config.Server
		confServer.SetHost(*confInput.Server.Host)
		confServer.SetPort(*confInput.Server.Port)
		confServer.SetOperatorJWT(*confInput.Server.OperatorJWT)
		confServer.SetAccountJWT(*confInput.Server.AccountJWT)
		confServer.SetSystemAccountJWT(*confInput.Server.SystemAccountJWT)

		confCommon, err := config.NewDefaultCommon()
		if err != nil {
			err := fmt.Errorf("cannot generate client configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		confOutput := config.Config{
			Common: confCommon,
			Server: &confServer,
		}
		_, _ = fmt.Fprintln(os.Stdout, confOutput.String())

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

	return nil
}

func validateInputConfiguration(conf *config.Config) error {
	if conf.Server == nil {
		return errors.New("server configuration is missing")
	}

	if conf.Server.Host == nil {
		return errors.New("host not configured")
	}

	if conf.Server.Port == nil {
		return errors.New("port not configured")
	}

	if conf.Server.OperatorJWT == nil {
		return errors.New("operator jwt not configured")
	}

	if conf.Server.OperatorKey == nil {
		return errors.New("operator key not configured")
	}

	if conf.Server.AccountJWT == nil {
		return errors.New("account jwt not configured")
	}

	if conf.Server.AccountKey == nil {
		return errors.New("account key not configured")
	}

	if conf.Server.SystemAccountJWT == nil {
		return errors.New("system account jwt not configured")
	}

	if conf.Server.SystemAccountKey == nil {
		return errors.New("system account key not configured")
	}

	return nil
}
