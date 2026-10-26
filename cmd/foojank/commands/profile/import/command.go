//nolint:staticcheck // preserve the package name
package import_

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/configdir"
	"github.com/foohq/foojank/internal/profile"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "import",
		Usage: "Import profiles from a file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      flags.File,
				Usage:     "set path to an input file",
				TakesFile: true,
			},
			&cli.BoolFlag{
				Name:  flags.Force,
				Usage: "overwrite existing profiles",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:          before,
		Action:          action,
		ShellComplete:   actions.CompleteFlags,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
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

func action(ctx context.Context, _ *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	profs := actions.GetProfilesFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	force, _ := conf.Bool(flags.Force)
	configDir, _ := conf.String(flags.ConfigDir)
	pth, _ := conf.String(flags.File)

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

		err = profs.Add(profName, profImport, profile.WithOverwrite(force))
		if err != nil {
			logger.ErrorContext(ctx, "Cannot import profile %q: %v", profName, err)
			return err
		}
	}

	err = configdir.UpdateProfilesJSON(configDir, profs)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create profile: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.File,
	} {
		switch opt {
		case flags.File:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("file not configured")
			}
		}
	}
	return nil
}
