package create

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/configdir"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/profile"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		ArgsUsage: "<name>",
		Usage:     "Create a new profile",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Os,
				Usage: "set OS environment variable",
			},
			&cli.StringFlag{
				Name:  flags.Arch,
				Usage: "set ARCH environment variable",
			},
			&cli.StringSliceFlag{
				Name:  flags.Feature,
				Usage: "set FEATURE environment variable",
			},
			&cli.StringSliceFlag{
				Name:  flags.Set,
				Usage: "set environment variable (format: key=value)",
			},
			&cli.StringFlag{
				Name:  flags.SourceDir,
				Usage: "set path to a source code directory",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:          before,
		Action:          action,
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

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	profs := actions.GetProfilesFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	configDir, _ := conf.String(flags.ConfigDir)
	sourceDir, _ := conf.String(flags.SourceDir)
	osVar, _ := conf.String(flags.Os)
	archVar, _ := conf.String(flags.Arch)
	featureVars, _ := conf.StringSlice(flags.Feature)
	setVars, _ := conf.StringSlice(flags.Set)

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	name := c.Args().First()

	if sourceDir != "" {
		var err error
		sourceDir, err = filepath.Abs(sourceDir)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get absolute path to source directory: %v", err)
			return err
		}
	}

	prof := profile.New()
	prof.SetSourceDir(sourceDir)

	if osVar != "" {
		prof.Set(profile.VarOS, profile.NewVar(osVar))
	}

	if archVar != "" {
		prof.Set(profile.VarArch, profile.NewVar(archVar))
	}

	if len(featureVars) > 0 {
		prof.Set(profile.VarFeatures, profile.NewVar(strings.Join(featureVars, ",")))
	}

	for k, v := range profile.ParseKVPairs(setVars) {
		prof.Set(k, v)
	}

	err := profs.Add(name, prof)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create profile: %v", err)
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
