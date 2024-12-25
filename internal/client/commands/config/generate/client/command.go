package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/log"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "client",
		ArgsUsage: "<config-file>",
		Usage:     "Generate client config from master/client config",
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
			err := fmt.Errorf("cannot parse configuration file: %v", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = validateInputConfiguration(input)
		if err != nil {
			err := fmt.Errorf("invalid input configuration file: %v", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		accountClaims, err := jwt.DecodeAccountClaims(input.Account.JWT)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot decode account JWT: %v", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		accountPubKey := accountClaims.Subject
		username := fmt.Sprintf("MG%s", nuid.Next())
		user, err := config.NewUserManager(username, accountPubKey, []byte(input.Account.SigningKeySeed))
		if err != nil {
			err := fmt.Errorf("cannot generate client configuration: %v", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		output := config.Config{
			Servers: input.Servers,
			Account: &config.Entity{
				JWT:            input.Account.JWT,
				SigningKeySeed: input.Account.SigningKeySeed,
			},
			User: user,
		}

		_, _ = fmt.Fprintln(os.Stdout, output.String())

		return nil
	}
}

func validateConfiguration(_ *config.Config) error {
	// TODO: validate LogLevel and NoColor
	return nil
}

func validateInputConfiguration(conf *config.Config) error {
	if conf.Servers == nil {
		return errors.New("servers not configured")
	}

	if conf.Account == nil {
		return errors.New("account not configured")
	}

	return nil
}
