package remove

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/configdir"
	"github.com/foohq/foojank/internal/foojank/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		ArgsUsage: "<name>",
		Usage:     "Remove a profile",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:       before,
		Action:       action,
		Aliases:      []string{"rm"},
		OnUsageError: actions.UsageError,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(os.Stderr, validateConfiguration)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.LoadProfiles(os.Stderr)(ctx, c)
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
	profs := actions.GetProfilesFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	configDir, _ := conf.String(flags.ConfigDir)

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	name := c.Args().First()

	err := profs.Delete(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot delete profile %q: %v", name, err)
		return err
	}

	err = configdir.UpdateProfilesJson(configDir, profs)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create profile: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
