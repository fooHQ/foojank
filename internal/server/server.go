package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/server/actions"
	"github.com/foohq/foojank/internal/server/flags"
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
				Value:   flags.DefaultConfig(),
				Aliases: []string{"c"},
			},
			&cli.StringFlag{
				Name:  flags.Host,
				Usage: "bind to host",
				Value: flags.DefaultHost,
			},
			&cli.IntFlag{
				Name:  flags.Port,
				Usage: "bind to port",
				Value: flags.DefaultPort,
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
			&cli.IntFlag{
				Name:  flags.LogLevel,
				Usage: "set log level",
				Value: flags.DefaultLogLevel,
			},
			&cli.BoolFlag{
				Name:  flags.NoColor,
				Usage: "disable color output",
				Value: flags.DefaultNoColor,
			},
		},
		Action:          action,
		DefaultCommand:  "start",
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	resolver := &server.MemAccResolver{}
	logger := actions.NewLogger(ctx, conf)
	return startAction(logger, conf, resolver)(ctx, c)
}

func startAction(logger *slog.Logger, conf *config.Config, resolver server.AccountResolver) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		// TODO: move validation inside a function!
		if conf.Host == nil {
			err := fmt.Errorf("invalid configuration: host not found")
			logger.Error(err.Error())
			return err
		}

		if conf.Port == nil {
			err := fmt.Errorf("invalid configuration: port not found")
			logger.Error(err.Error())
			return err
		}

		operator := conf.Operator
		if operator == nil {
			err := fmt.Errorf("invalid configuration: no operator found")
			logger.Error(err.Error())
			return err
		}

		account := conf.Account
		if account == nil {
			err := fmt.Errorf("invalid configuration: no account found")
			logger.Error(err.Error())
			return err
		}

		systemAccount := conf.SystemAccount
		if account == nil {
			err := fmt.Errorf("invalid configuration: no system account found")
			logger.Error(err.Error())
			return err
		}

		var preloadOperators []*jwt.OperatorClaims
		for _, operatorJWT := range []string{operator.JWT} {
			claims, err := jwt.DecodeOperatorClaims(operatorJWT)
			if err != nil {
				err := fmt.Errorf("invalid configuration: cannot decode operator JWT: %v", err)
				logger.Error(err.Error())
				return err
			}

			preloadOperators = append(preloadOperators, claims)
		}

		for _, accountJWT := range []string{account.JWT, systemAccount.JWT} {
			claims, err := jwt.DecodeAccountClaims(accountJWT)
			if err != nil {
				err := fmt.Errorf("invalid configuration: cannot decode account JWT: %v", err)
				logger.Error(err.Error())
				return err
			}

			accountPubKey := claims.Subject
			err = resolver.Store(accountPubKey, accountJWT)
			if err != nil {
				err := fmt.Errorf("cannot store account in the resolver: %v", err)
				logger.Error(err.Error())
				return err
			}
		}

		claims, err := jwt.DecodeAccountClaims(systemAccount.JWT)
		if err != nil {
			err := fmt.Errorf("invalid configuration: cannot decode account JWT: %v", err)
			logger.Error(err.Error())
			return err
		}

		systemAccountPubKey := claims.Subject
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
