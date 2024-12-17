package actions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/server/flags"
)

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
			Host:     &flags.DefaultHost,
			Port:     &flags.DefaultPort,
			LogLevel: &flags.DefaultLogLevel,
			NoColor:  &flags.DefaultNoColor,
		}
	}

	if c.IsSet(flags.Host) {
		v := c.String(flags.Host)
		conf.Host = &v
	}

	if c.IsSet(flags.Port) {
		v := c.Int(flags.Port)
		conf.Port = &v
	}

	if c.IsSet(flags.OperatorJWT) {
		v := c.String(flags.OperatorJWT)
		conf.Operator = &config.Entity{
			JWT: v,
		}
	}

	if c.IsSet(flags.AccountJWT) {
		v := c.String(flags.AccountJWT)
		conf.Account = &config.Entity{
			JWT: v,
		}
	}

	if c.IsSet(flags.SystemAccountJWT) {
		v := c.String(flags.SystemAccountJWT)
		conf.SystemAccount = &config.Entity{
			JWT: v,
		}
	}

	if c.IsSet(flags.LogLevel) {
		v := c.Int(flags.LogLevel)
		conf.LogLevel = &v
	}

	if c.IsSet(flags.NoColor) {
		v := c.Bool(flags.NoColor)
		conf.NoColor = &v
	}

	return conf, nil
}
