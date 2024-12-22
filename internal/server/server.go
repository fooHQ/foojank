package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/server/actions"
	"github.com/foohq/foojank/internal/server/flags"
	"github.com/foohq/foojank/internal/server/log"
)

func New() *cli.Command {
	return &cli.Command{
		Name:    "foojank",
		Usage:   "A foojank server",
		Version: foojank.Version(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flags.Config,
				Usage:   "path to a configuration file",
				Value:   config.DefaultServerConfigPath(),
				Aliases: []string{"c"},
			},
			&cli.StringFlag{
				Name:  flags.Host,
				Usage: "bind to host",
				Value: config.DefaultHost,
			},
			&cli.IntFlag{
				Name:  flags.Port,
				Usage: "bind to port",
				Value: config.DefaultPort,
			},
			&cli.StringFlag{
				Name:  flags.OperatorJWT,
				Usage: "operator JWT token",
			},
			&cli.StringFlag{
				Name:  flags.SystemAccountJWT,
				Usage: "system account JWT token",
			},
			&cli.StringFlag{
				Name:  flags.AccountJWT,
				Usage: "account JWT token",
			},
			&cli.StringFlag{
				Name:  flags.LogLevel,
				Usage: "set log level",
				Value: config.DefaultLogLevel,
			},
			&cli.BoolFlag{
				Name:  flags.NoColor,
				Usage: "disable color output",
				Value: config.DefaultNoColor,
			},
		},
		Action:          action,
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
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
	resolver := &server.MemAccResolver{}
	return startAction(logger, conf, resolver)(ctx, c)
}

func startAction(logger *slog.Logger, conf *config.Config, resolver server.AccountResolver) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		preloadOperators, err := decodeOperatorClaims(conf.Operator.JWT)
		if err != nil {
			err := fmt.Errorf("invalid configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		configuredAccounts := []string{
			conf.Account.JWT,
			conf.SystemAccount.JWT,
		}
		preloadAccounts, err := decodeAccountClaims(configuredAccounts...)
		if err != nil {
			err := fmt.Errorf("invalid configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		for i, claims := range preloadAccounts {
			accountPubKey := claims.Subject
			accountJWT := configuredAccounts[i]
			err = resolver.Store(accountPubKey, accountJWT)
			if err != nil {
				err := fmt.Errorf("cannot store account in the resolver: %v", err)
				logger.Error(err.Error())
				return err
			}
		}

		// This is a footgun waiting to hurt someone.
		// System account must always be defined as the last in the decodeAccountClaims.
		systemAccountPubKey := preloadAccounts[len(preloadAccounts)-1].Subject
		opts := &server.Options{
			Host:             *conf.Host,
			Port:             int(*conf.Port),
			SystemAccount:    systemAccountPubKey,
			JetStream:        true,
			AccountResolver:  resolver,
			TrustedOperators: preloadOperators,
		}
		s, err := server.NewServer(opts)
		if err != nil {
			err := fmt.Errorf("cannot start a server: %v", err)
			logger.Error(err.Error())
			return err
		}
		s.ConfigureLogger()

		go func() {
			err := server.Run(s)
			if err != nil {
				logger.Error(err.Error())
			}
		}()

		s.WaitForShutdown()

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.Host == nil {
		return fmt.Errorf("host not found")
	}

	if conf.Port == nil {
		return fmt.Errorf("port not found")
	}

	if conf.Operator == nil {
		return fmt.Errorf("no operator found")
	}

	if conf.Account == nil {
		return fmt.Errorf("no account found")
	}

	if conf.SystemAccount == nil {
		return fmt.Errorf("no system account found")
	}

	return nil
}

func decodeOperatorClaims(operatorJWTs ...string) ([]*jwt.OperatorClaims, error) {
	var result []*jwt.OperatorClaims
	for _, operatorJWT := range operatorJWTs {
		claims, err := jwt.DecodeOperatorClaims(operatorJWT)
		if err != nil {
			err := fmt.Errorf("cannot decode operator JWT: %v", err)
			return nil, err
		}

		result = append(result, claims)
	}
	return result, nil
}

func decodeAccountClaims(accountJWTs ...string) ([]*jwt.AccountClaims, error) {
	var result []*jwt.AccountClaims
	for _, accountJWT := range accountJWTs {
		claims, err := jwt.DecodeAccountClaims(accountJWT)
		if err != nil {
			err := fmt.Errorf("cannot decode account JWT: %v", err)
			return nil, err
		}

		result = append(result, claims)
	}
	return result, nil
}
