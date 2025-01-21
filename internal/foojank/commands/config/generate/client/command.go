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

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
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
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
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
			err := fmt.Errorf("cannot parse configuration file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = validateInputConfiguration(confInput)
		if err != nil {
			err := fmt.Errorf("invalid input configuration file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		accountClaims, err := jwt.DecodeAccountClaims(*confInput.Client.AccountJWT)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot decode account JWT: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		accountPubKey := accountClaims.Subject
		username := fmt.Sprintf("MG%s", nuid.Next())
		user, err := auth.NewUserManager(username, accountPubKey, []byte(*confInput.Client.AccountKey))
		if err != nil {
			err := fmt.Errorf("cannot generate client configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		var confClient config.Client
		confClient.SetServer(confInput.Client.Server)
		confClient.SetUserJWT(user.JWT)
		confClient.SetUserKey(user.Key)
		confClient.SetAccountJWT(*confInput.Client.AccountJWT)
		confClient.SetAccountKey(*confInput.Client.AccountKey)

		confCommon, err := config.NewDefaultCommon()
		if err != nil {
			err := fmt.Errorf("cannot generate client configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		confOutput := config.Config{
			Common: confCommon,
			Client: &confClient,
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
	if conf.Client == nil {
		return errors.New("client configuration is missing")
	}

	if len(conf.Client.Server) == 0 {
		return errors.New("server not configured")
	}

	if conf.Client.AccountJWT == nil {
		return errors.New("account jwt not configured")
	}

	if conf.Client.AccountKey == nil {
		return errors.New("account key not configured")
	}

	return nil
}
