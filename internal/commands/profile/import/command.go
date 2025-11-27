package _import

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/configdir"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/profile"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "import",
		ArgsUsage: "<file>",
		Usage:     "Import profiles from a file",
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

	pth := c.Args().First()

	profsImport, err := profile.ParseFile(pth)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot parse profiles in %q: %v", pth, err)
		return err
	}

	for _, profName := range profsImport.List() {
		profImport, err := profsImport.Get(profName)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot find profile %q in file %q: %v", profName, pth, err)
			return err
		}

		err = profs.Add(profName, profImport)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot import profile %q: %v", profName, err)
			return err
		}
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
