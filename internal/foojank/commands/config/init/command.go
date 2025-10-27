package init

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:         "init",
		Usage:        "Initialize configuration directory",
		Action:       action,
		OnUsageError: actions.UsageError,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	confFlags, err := config.ParseFlags(c.FlagNames(), func(name string) (any, bool) {
		return c.Value(name), c.IsSet(name)
	})
	if err != nil {
		log.Error(ctx, "Cannot parse command options: %v", err)
		return err
	}

	configDir, isSet := confFlags.String(flags.ConfigDir)
	if !isSet {
		dir, err := actions.FindConfigDir(".")
		if err == nil {
			configDir = dir
		} else {
			configDir = "."
		}
	}

	configDir, err = filepath.Abs(configDir)
	if err != nil {
		log.Error(ctx, "Cannot resolve configuration directory: %v", err)
		return err
	}

	isConfigDir, err := actions.IsConfigDir(configDir)
	if err != nil {
		log.Error(ctx, "Cannot initialize configuration directory %q: %v", configDir, errors.Unwrap(err))
		return err
	}

	if isConfigDir {
		log.Info(ctx, "Configuration directory has already been initialized in %q", configDir)
		return nil
	}

	err = actions.InitConfigDir(configDir)
	if err != nil {
		log.Error(ctx, "Cannot initialize configuration directory %q: %v", configDir, errors.Unwrap(err))
		return err
	}

	log.Info(ctx, "Initialized empty configuration directory in %q", configDir)

	return nil
}
