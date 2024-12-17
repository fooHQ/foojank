package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config"
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
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)

	return generateAction(logger)(ctx, c)
}

func generateAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		input, err := config.ParseFile(c.Args().First())
		if err != nil {
			err := fmt.Errorf("cannot parse configuration file: %v", err)
			logger.Error(err.Error())
			return err
		}

		operator := input.Operator
		if operator == nil {
			err := fmt.Errorf("cannot generate server configuration: no operator found")
			logger.Error(err.Error())
			return err
		}

		account := input.Account
		if account == nil {
			err := fmt.Errorf("cannot generate server configuration: no account found")
			logger.Error(err.Error())
			return err
		}

		systemAccount := input.SystemAccount
		if account == nil {
			err := fmt.Errorf("cannot generate server configuration: no system account found")
			logger.Error(err.Error())
			return err
		}

		output := config.Config{
			Host: input.Host,
			Port: input.Port,
			Operator: &config.Entity{
				JWT: operator.JWT,
			},
			Account: &config.Entity{
				JWT: account.JWT,
			},
			SystemAccount: &config.Entity{
				JWT: systemAccount.JWT,
			},
		}

		_, _ = fmt.Fprintln(os.Stdout, output.String())

		return nil
	}
}
