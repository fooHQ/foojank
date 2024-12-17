package start

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/server/actions"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:   "start",
		Usage:  "Start server",
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)
	return startAction(logger, conf)(ctx, c)
}

func startAction(logger *slog.Logger, conf *config.Config) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		// TODO: configurable directory!
		// TODO: we can probably use memory resolver?
		resolver, err := server.NewDirAccResolver("/tmp/nats-jwt", 0, 0, 1, server.FetchTimeout(2*time.Second))
		if err != nil {
			err := fmt.Errorf("cannot create account resolver: %v", err)
			logger.Error(err.Error())
			return err
		}

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
