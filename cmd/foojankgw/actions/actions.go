package actions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/deepnoodle-ai/wonton/tty"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/auth"

	"github.com/foohq/foojank/internal/authdir"

	"github.com/foohq/foojank/internal/config"

	"github.com/foohq/foojank/internal/log"

	"github.com/foohq/foojank/cmd/foojankgw/flags"
)

func LoadConfig(w io.Writer, validateFn func(conf *config.Config) error) cli.BeforeFunc {
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		newCtx, err := loadConfig(w, validateFn)(ctx, c)
		// urfave skips ShellComplete when Before fails. Flag names do not need a
		// working server or a project config, so keep completion running.
		if err != nil && IsShellCompletion() {
			return setConfigToContext(ctx, flagsOnlyConfig(c)), nil
		}
		return newCtx, err
	}
}

func loadConfig(w io.Writer, validateFn func(conf *config.Config) error) cli.BeforeFunc {
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		if IsShellCompletion() {
			w = io.Discard
		}
		confFlags, err := config.ParseFlags(c.FlagNames(), func(name string) (any, bool) {
			return c.Value(name), c.IsSet(name)
		})
		if err != nil {
			_, _ = fmt.Fprintf(w, "%s: cannot parse command options: %v\n", c.FullName(), err)
			return ctx, err
		}

		confDefs := config.NewWithOptions(map[string]string{
			flags.Format:  "ascii",
			flags.NoColor: "false",
		})

		confs := []*config.Config{
			confDefs,
			confFlags,
		}

		// If the output is not a TTY, disable color output.
		if !tty.IsTerminal(os.Stdout) {
			confs = append(confs, config.NewWithOptions(map[string]string{
				flags.NoColor: "true",
			}))
		}

		conf := config.Merge(confs...)

		if !IsShellCompletion() {
			err = validateFn(conf)
			if err != nil {
				_, _ = fmt.Fprintf(w, "%s: invalid configuration: %v\n", c.FullName(), err)
				return ctx, err
			}
		}

		return setConfigToContext(ctx, conf), nil
	}
}

func flagsOnlyConfig(c *cli.Command) *config.Config {
	confFlags, err := config.ParseFlags(c.FlagNames(), func(name string) (any, bool) {
		return c.Value(name), c.IsSet(name)
	})
	if err != nil {
		confFlags = config.NewWithOptions(nil)
	}
	return config.Merge(config.NewWithOptions(map[string]string{
		flags.Format:  "ascii",
		flags.NoColor: "false",
	}), confFlags)
}

func SetupLogger(_ io.Writer) cli.BeforeFunc {
	return func(ctx context.Context, _ *cli.Command) (context.Context, error) {
		conf := GetConfigFromContext(ctx)

		noColor, ok := conf.Bool(flags.NoColor)
		if !ok {
			noColor = false
		}

		logger := log.NewLogger(log.LevelInfo, noColor)
		return setLoggerToContext(ctx, logger), nil
	}
}

func SetupServer(_ io.Writer) cli.BeforeFunc {
	return func(ctx context.Context, _ *cli.Command) (context.Context, error) {
		// urfave skips ShellComplete when Before fails. Flag names do not need
		// a live server connection.
		if IsShellCompletion() {
			return ctx, nil
		}

		conf := GetConfigFromContext(ctx)
		logger := GetLoggerFromContext(ctx)

		serverURL, _ := conf.String(flags.ServerURL)
		serverCert, _ := conf.String(flags.ServerCertificate)
		credsName, _ := conf.String(flags.Credential)

		user, err := authdir.GetUser(credsName)
		if err != nil {
			if errors.Is(err, authdir.ErrUserNotFound) {
				err = fmt.Errorf("%q not found", credsName)
			}
			logger.ErrorContext(ctx, "Cannot read credential: %v", err)
			return ctx, err
		}

		if auth.IsDummyUserJWT(user.Claims()) {
			err := fmt.Errorf("%q has not been activated", user.Claims().Name)
			logger.ErrorContext(ctx, "Cannot use credential: %v", err)
			return ctx, err
		}

		userJWT := user.JWT()
		userSeed, err := user.KeyPair().Seed()
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get credential seed: %v", err)
			return ctx, err
		}

		srv, err := server.New(
			ctx,
			[]string{serverURL},
			userJWT,
			string(userSeed),
			server.WithServerCert(serverCert),
		)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
			return ctx, err
		}

		return setServerToContext(ctx, srv), nil
	}
}

func UsageError(_ context.Context, c *cli.Command, err error, _ bool) error {
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	return nil
}

func CommandNotFound(_ context.Context, c *cli.Command, s string) {
	err := fmt.Errorf("%q is not a valid command", s)
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", c.FullName(), err.Error())
	os.Exit(1)
}

type contextKey string

var (
	configKey contextKey = "foojank:config"
	loggerKey contextKey = "foojank:logger"
	serverKey contextKey = "foojank:server"
)

func GetConfigFromContext(ctx context.Context) *config.Config {
	// The function will panic if a context key is not found, that's intended to catch bugs early.
	conf := ctx.Value(configKey).(*config.Config)
	return conf
}

func GetLoggerFromContext(ctx context.Context) *log.Logger {
	logger := ctx.Value(loggerKey).(*log.Logger)
	return logger
}

func GetServerFromContext(ctx context.Context) *server.Client {
	srv := ctx.Value(serverKey).(*server.Client)
	return srv
}

func setConfigToContext(ctx context.Context, conf *config.Config) context.Context {
	return context.WithValue(ctx, configKey, conf)
}

func setLoggerToContext(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func setServerToContext(ctx context.Context, srv *server.Client) context.Context {
	return context.WithValue(ctx, serverKey, srv)
}
