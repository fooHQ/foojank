package start

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/flags"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagHost             = "host"
	FlagPort             = "port"
	FlagOperatorJWT      = "operator-jwt"
	FlagSystemAccountJWT = "system-account-jwt"
	FlagAccountJWT       = flags.AccountJWT
	FlagDataDir          = flags.DataDir
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "Start a server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  FlagHost,
				Usage: "bind to host",
			},
			&cli.IntFlag{
				Name:  FlagPort,
				Usage: "bind to port",
			},
			&cli.StringFlag{
				Name:  FlagOperatorJWT,
				Usage: "set operator JWT token",
			},
			&cli.StringFlag{
				Name:  FlagAccountJWT,
				Usage: "set account JWT token",
			},
			&cli.StringFlag{
				Name:  FlagSystemAccountJWT,
				Usage: "set system account JWT token",
			},
			&cli.StringFlag{
				Name:  FlagDataDir,
				Usage: "set path to a data directory",
			},
		},
		Action:  action,
		Aliases: []string{"exec"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
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
	resolver := &server.MemAccResolver{}
	return startAction(logger, conf, resolver)(ctx, c)
}

func startAction(logger *slog.Logger, conf *config.Config, resolver server.AccountResolver) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		preloadOperators, err := decodeOperatorClaims(*conf.Server.OperatorJWT)
		if err != nil {
			err := fmt.Errorf("invalid configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		configuredAccounts := []string{
			*conf.Server.AccountJWT,
			*conf.Server.SystemAccountJWT,
		}
		preloadAccounts, err := decodeAccountClaims(configuredAccounts...)
		if err != nil {
			err := fmt.Errorf("invalid configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		for i, claims := range preloadAccounts {
			accountPubKey := claims.Subject
			accountJWT := configuredAccounts[i]
			err = resolver.Store(accountPubKey, accountJWT)
			if err != nil {
				err := fmt.Errorf("cannot store account in the resolver: %w", err)
				logger.ErrorContext(ctx, err.Error())
				return err
			}
		}

		// This is a footgun waiting to hurt someone.
		// System account must always be defined as the last in the configuredAccounts.
		systemAccountPubKey := preloadAccounts[len(preloadAccounts)-1].Subject
		opts := &server.Options{
			Host: "localhost",
			Port: 4222,
			Websocket: server.WebsocketOpts{
				Host:  *conf.Server.Host,
				Port:  int(*conf.Server.Port),
				NoTLS: true,
				// TODO: consider enabling the compression!
				Compression: false,
			},
			SystemAccount:    systemAccountPubKey,
			JetStream:        true,
			AccountResolver:  resolver,
			TrustedOperators: preloadOperators,
			StoreDir:         *conf.DataDir,
		}
		s, err := server.NewServer(opts)
		if err != nil {
			err := fmt.Errorf("cannot start a server: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}
		s.ConfigureLogger()

		go func() {
			err := server.Run(s)
			if err != nil {
				logger.ErrorContext(ctx, err.Error())
			}
		}()

		s.WaitForShutdown()

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.DataDir == nil {
		return errors.New("data directory not configured")
	}

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

	if conf.Server.AccountJWT == nil {
		return errors.New("account jwt not configured")
	}

	if conf.Server.SystemAccountJWT == nil {
		return errors.New("system account jwt not configured")
	}

	return nil
}

func decodeOperatorClaims(operatorJWTs ...string) ([]*jwt.OperatorClaims, error) {
	var result []*jwt.OperatorClaims
	for _, operatorJWT := range operatorJWTs {
		claims, err := jwt.DecodeOperatorClaims(operatorJWT)
		if err != nil {
			err := fmt.Errorf("cannot decode operator JWT: %w", err)
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
			err := fmt.Errorf("cannot decode account JWT: %w", err)
			return nil, err
		}

		result = append(result, claims)
	}
	return result, nil
}
