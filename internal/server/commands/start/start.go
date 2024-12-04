package start

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/server/actions"
	"github.com/foohq/foojank/internal/server/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:   "start",
		Usage:  "Start server",
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	logger := actions.NewLogger(ctx, c)
	return startAction(logger)(ctx, c)
}

func startAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		// TODO: load from --config
		conf, err := config.Parse(c.Args().First())
		if err != nil {
			err = fmt.Errorf("cannot parse configuration file: %v", err)
			logger.Error(err.Error())
			return err
		}

		// TODO: configurable directory!
		resolver, err := server.NewDirAccResolver("/tmp/nats-jwt", 0, 0, 1, server.FetchTimeout(2*time.Second))
		if err != nil {
			err := fmt.Errorf("cannot create account resolver: %v", err)
			logger.Error(err.Error())
			return err
		}

		preloadAccounts := make(map[string]string)
		preloadAccounts[conf.SystemAccount.Key] = conf.SystemAccount.JWT
		for _, account := range conf.Accounts {
			preloadAccounts[account.Key] = account.JWT
		}

		for accountKey, accountJWT := range preloadAccounts {
			err = resolver.Store(accountKey, accountJWT)
			if err != nil {
				err := fmt.Errorf("cannot store account: %v", err)
				logger.Error(err.Error())
				return err
			}
		}

		var preloadOperators []*jwt.OperatorClaims
		for _, operator := range conf.Operators {
			operatorClaims, err := jwt.DecodeOperatorClaims(operator.JWT)
			if err != nil {
				err := fmt.Errorf("cannot decode operator JWT: %v", err)
				logger.Error(err.Error())
				return err
			}

			preloadOperators = append(preloadOperators, operatorClaims)
		}

		opts := &server.Options{
			Host:             "127.0.0.1",
			Port:             4222,
			SystemAccount:    conf.SystemAccount.Key,
			AccountResolver:  resolver,
			TrustedOperators: preloadOperators,
		}
		s, err := server.NewServer(opts)
		if err != nil {
			err := fmt.Errorf("cannot create server: %v", err)
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
