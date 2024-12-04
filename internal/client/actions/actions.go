package actions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/nats-io/nats.go"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/config"
	"github.com/foohq/foojank/internal/client/flags"
)

func NewConfig(ctx context.Context, c *cli.Command) (*config.Config, error) {
	file := c.String(flags.Config)
	conf, err := config.Parse(file)
	if err != nil {
		if errors.Is(err, config.ErrParserError) {
			err = fmt.Errorf("cannot parse configuration file '%s': %v", file, err)
			_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err)
			return nil, err
		}
		conf = &config.Config{}
	}

	switch {
	case c.IsSet(flags.Server):
		conf.Servers = c.StringSlice(flags.Server)
	case c.IsSet(flags.UserJWT):
		conf.User.JWT = c.String(flags.UserJWT)
	case c.IsSet(flags.UserNkey):
		conf.User.Key = c.String(flags.UserNkey)
	case c.IsSet(flags.LogLevel):
		conf.LogLevel = c.Int(flags.LogLevel)
	case c.IsSet(flags.NoColor):
		conf.NoColor = c.Bool(flags.NoColor)
	}

	err = conf.Validate()
	if err != nil {
		err = fmt.Errorf("invalid configuration: %v", err)
		_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err)
		return nil, err
	}

	return conf, nil
}

func CommandNotFound(ctx context.Context, c *cli.Command, s string) {
	err := fmt.Errorf("command '%s' does not exist", s)
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	os.Exit(1)
}

func NewLogger(ctx context.Context, conf *config.Config) *slog.Logger {
	return slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:     slog.Level(conf.LogLevel),
		NoColor:   conf.NoColor,
		AddSource: true,
	}))
}

func NewServerConnection(ctx context.Context, conf *config.Config, logger *slog.Logger) (*nats.Conn, error) {
	servers := strings.Join(conf.Servers, ",")
	nc, err := nats.Connect(
		servers,
		nats.UserJWTAndSeed(conf.User.JWT, conf.User.Key),
		nats.MaxReconnects(-1),
		nats.ConnectHandler(func(nc *nats.Conn) {
			logger.Debug("connected to the server")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("reconnected to the server")
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			err = fmt.Errorf("disconnected from the server: %v", err)
			logger.Warn(err.Error())
		}),
		/*nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
			// TODO: set better error message
			logger.Warn("NATS error ", "error", err, "server", server)
		}),*/
	)
	if err != nil {
		err = fmt.Errorf("cannot connect to the server: %v", err)
		logger.Error(err.Error())
		return nil, err
	}

	return nc, nil
}
