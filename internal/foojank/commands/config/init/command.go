package init

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/configdir"
	"github.com/foohq/foojank/internal/foojank/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize configuration directory",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:       before,
		Action:       action,
		OnUsageError: actions.UsageError,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadFlags(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	configDir, isSet := conf.String(flags.ConfigDir)
	if !isSet {
		dir, err := configdir.Search(".")
		if err == nil {
			configDir = dir
		} else {
			configDir = "."
		}
	}

	configDir, err := filepath.Abs(configDir)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot resolve configuration directory: %v", err)
		return err
	}

	isConfigDir, err := configdir.IsConfigDir(configDir)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot initialize configuration directory %q: %v", configDir, errors.Unwrap(err))
		return err
	}

	if isConfigDir {
		logger.InfoContext(ctx, "Configuration directory has already been initialized in %q", configDir)
		return nil
	}

	err = configdir.Init(configDir)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot initialize configuration directory %q: %v", configDir, errors.Unwrap(err))
		return err
	}

	logger.InfoContext(ctx, "Initialized empty configuration directory in %q", configDir)

	return nil
}
