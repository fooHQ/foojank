package actions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/configdir"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
)

func LoadConfig(w io.Writer, validateFn func(conf *config.Config) error) cli.BeforeFunc {
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		confFlags, err := config.ParseFlags(c.FlagNames(), func(name string) (any, bool) {
			return c.Value(name), c.IsSet(name)
		})
		if err != nil {
			err = fmt.Errorf("cannot parse command options: %w", err)
			return nil, err
		}

		configDir, isSet := confFlags.String(flags.ConfigDir)
		if !isSet {
			dir, err := configdir.Search(".")
			if err != nil {
				return nil, errors.New("configuration directory not found in the current directory (or any of the parent directories)")
			}

			configDir = dir
		}

		isConfigDir, err := configdir.IsConfigDir(configDir)
		if err != nil {
			return nil, err
		}

		if !isConfigDir {
			err = fmt.Errorf("configuration directory not found in %q", configDir)
			return nil, err
		}

		confFile, err := configdir.ParseConfigJson(configDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = fmt.Errorf("configuration directory not found in %q", configDir)
			} else {
				err = fmt.Errorf("cannot parse config file: %w", err)
			}
			return nil, err
		}

		conf := config.Merge(newDefaultConfig(), confFile, confFlags)

		err = validateFn(conf)
		if err != nil {
			_, _ = fmt.Fprintf(w, "%s: invalid configuration: %v\n", c.FullName(), err)
			return ctx, err
		}

		return setConfigToContext(ctx, conf), nil
	}
}

func LoadFlags() cli.BeforeFunc {
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		conf, err := config.ParseFlags(c.FlagNames(), func(name string) (any, bool) {
			return c.Value(name), c.IsSet(name)
		})
		if err != nil {
			err = fmt.Errorf("cannot parse command options: %w", err)
			return ctx, err
		}

		return setConfigToContext(ctx, conf), nil
	}
}

func newDefaultConfig() *config.Config {
	opts := map[string]any{
		flags.Format: "table",
	}

	if true { // TODO: check if output is tty!
		opts[flags.NoColor] = false
	}

	return config.NewWithOptions(opts)
}

func SetupLogger() cli.BeforeFunc {
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		conf := GetConfigFromContext(ctx)

		noColor, ok := conf.Bool(flags.NoColor)
		if !ok {
			noColor = false
		}

		logger := log.NewLogger(log.LevelInfo, noColor)
		return setLoggerToContext(ctx, logger), nil
	}
}

func UsageError(ctx context.Context, c *cli.Command, err error, _ bool) error {
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

func setConfigToContext(ctx context.Context, conf *config.Config) context.Context {
	return context.WithValue(ctx, configKey, conf)
}

func setLoggerToContext(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
