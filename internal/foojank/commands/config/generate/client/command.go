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
		ArgsUsage: "<file>",
		Usage:     "Generate client configuration from server/client configuration file",
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

		configPth := c.Args().First()
		confInput, err := config.ParseFile(configPth)
		if err != nil {
			err := fmt.Errorf("cannot parse configuration file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		confClient, err := fromConfig(confInput)
		if err != nil {
			err := fmt.Errorf("cannot generate client configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		err = validateClientConfiguration(confClient)
		if err != nil {
			err := fmt.Errorf("invalid input configuration file: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		accountClaims, err := jwt.DecodeAccountClaims(*confClient.AccountJWT)
		if err != nil {
			err := fmt.Errorf("cannot decode account JWT: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		accountPubKey := accountClaims.Subject
		username := fmt.Sprintf("MG%s", nuid.Next())
		user, err := auth.NewUserManager(username, accountPubKey, []byte(*confClient.AccountKey))
		if err != nil {
			err := fmt.Errorf("cannot generate client configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		confClient.SetUserJWT(user.JWT)
		confClient.SetUserKey(user.Key)

		confCommon, err := config.NewDefaultCommon()
		if err != nil {
			err := fmt.Errorf("cannot generate client configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		confOutput := config.Config{
			Common: confCommon,
			Client: confClient,
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

func fromConfig(conf *config.Config) (*config.Client, error) {
	var confs []*config.Client

	confDef, err := config.NewDefaultClient()
	if err != nil {
		return nil, err
	}

	confs = append(confs, confDef)

	if conf.Client != nil {
		confs = append(confs, conf.Client)
	}

	if conf.Server != nil {
		confs = append(confs, convertServerToClientConfig(conf.Server))
	}

	return config.MergeClient(confs...), nil
}

func validateClientConfiguration(conf *config.Client) error {
	if conf.AccountJWT == nil {
		return errors.New("account jwt not configured")
	}

	if conf.AccountKey == nil {
		return errors.New("account key not configured")
	}

	if conf.TLSCACertificate == nil {
		return errors.New("tls ca certificate not configured")
	}

	return nil
}

func convertServerToClientConfig(conf *config.Server) *config.Client {
	return &config.Client{
		AccountJWT:       conf.AccountJWT,
		AccountKey:       conf.AccountKey,
		TLSCACertificate: conf.TLSCertificate,
	}
}
