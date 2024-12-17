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

	"github.com/foohq/foojank/internal/client/flags"
	"github.com/foohq/foojank/internal/config"
)

func NewConfig(ctx context.Context, c *cli.Command) (*config.Config, error) {
	file := c.String(flags.Config)
	conf, err := config.ParseFile(file)
	if err != nil {
		if errors.Is(err, config.ErrParserError) {
			err = fmt.Errorf("cannot parse configuration file '%s': %v", file, err)
			_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err)
			return nil, err
		} else if !errors.Is(err, os.ErrNotExist) {
			err = fmt.Errorf("cannot open configuration file '%s': %v", file, err)
			_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err)
			return nil, err
		}

		// File does not exist, fallthrough.
		conf = &config.Config{
			Servers:  flags.DefaultServer,
			LogLevel: &flags.DefaultLogLevel,
			NoColor:  &flags.DefaultNoColor,
		}
	}

	if c.IsSet(flags.Server) {
		conf.Servers = c.StringSlice(flags.Server)
	}

	if c.IsSet(flags.UserJWT) {
		if conf.User == nil {
			conf.User = &config.Entity{}
		}
		conf.User.JWT = c.String(flags.UserJWT)
	}

	if c.IsSet(flags.UserKey) {
		if conf.User == nil {
			conf.User = &config.Entity{}
		}
		conf.User.KeySeed = c.String(flags.UserKey)
	}

	if c.IsSet(flags.AccountJWT) {
		if conf.Account == nil {
			conf.Account = &config.Entity{}
		}
		conf.Account.JWT = c.String(flags.AccountJWT)
	}

	if c.IsSet(flags.AccountSigningKey) {
		if conf.Account == nil {
			conf.Account = &config.Entity{}
		}
		conf.Account.SigningKeySeed = c.String(flags.AccountSigningKey)
	}

	if c.IsSet(flags.LogLevel) {
		v := c.Int(flags.LogLevel)
		conf.LogLevel = &v
	}

	if c.IsSet(flags.NoColor) {
		v := c.Bool(flags.NoColor)
		conf.NoColor = &v
	}

	if c.IsSet(flags.Codebase) {
		v := c.String(flags.Codebase)
		conf.Codebase = &v
	}

	return conf, nil
}

func CommandNotFound(ctx context.Context, c *cli.Command, s string) {
	err := fmt.Errorf("command '%s' does not exist", s)
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	os.Exit(1)
}

func NewLogger(ctx context.Context, conf *config.Config) *slog.Logger {
	logLevel := slog.LevelInfo
	if conf.LogLevel != nil {
		logLevel = slog.Level(*conf.LogLevel)
	}

	noColor := false
	if conf.NoColor != nil {
		noColor = *conf.NoColor
	}

	return slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:     logLevel,
		NoColor:   noColor,
		AddSource: logLevel == slog.LevelDebug,
	}))
}

func NewServerConnection(ctx context.Context, conf *config.Config, logger *slog.Logger) (*nats.Conn, error) {
	servers := strings.Join(conf.Servers, ",")
	user := conf.User
	if user == nil {
		err := fmt.Errorf("user configuration is missing")
		logger.Error(err.Error())
		return nil, err
	}

	nc, err := nats.Connect(
		servers,
		nats.UserJWTAndSeed(user.JWT, user.KeySeed),
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
