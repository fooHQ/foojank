package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config"
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
	conf, err := actions.NewConfig(ctx, c, validateConfiguration)
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
		input, err := config.ParseFile(confFile, true)
		if err != nil {
			err := fmt.Errorf("cannot parse master configuration file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = validateInputConfiguration(input)
		if err != nil {
			err := fmt.Errorf("invalid master configuration file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		output := config.Config{
			Host: input.Host,
			Port: input.Port,
			Operator: &config.Entity{
				JWT: input.Operator.JWT,
			},
			Account: &config.Entity{
				JWT: input.Account.JWT,
			},
			SystemAccount: &config.Entity{
				JWT: input.SystemAccount.JWT,
			},
		}

		_, _ = fmt.Fprintln(os.Stdout, output.String())

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
	if conf.Host == nil {
		return errors.New("host not configured")
	}

	if conf.Port == nil {
		return errors.New("port not configured")
	}

	if conf.Operator == nil {
		return errors.New("operator not configured")
	}

	if conf.Account == nil {
		return errors.New("account not configured")
	}

	if conf.SystemAccount == nil {
		return errors.New("system account not configured")
	}

	return nil
}
