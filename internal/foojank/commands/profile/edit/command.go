package edit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/configdir"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/profile"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		ArgsUsage: "<name>",
		Usage:     "Edit a profile",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Name,
				Usage: "set new name",
			},
			&cli.StringFlag{
				Name:  flags.SourceDir,
				Usage: "set path to a source code directory",
			},
			&cli.StringSliceFlag{
				Name:  flags.Set,
				Usage: "set environment variable (format: KEY=value)",
			},
			&cli.StringSliceFlag{
				Name:  flags.Unset,
				Usage: "unset environment variable (format: KEY)",
			},
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
	sourceDir, _ := conf.String(flags.SourceDir)
	setVars, _ := conf.StringSlice(flags.Set)
	unsetVars, _ := conf.StringSlice(flags.Unset)
	newName, _ := conf.String(flags.Name)

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	name := c.Args().First()

	prof, err := profs.Get(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get profile %q: %v", name, err)
		return err
	}

	if sourceDir != "" {
		var err error
		sourceDir, err = filepath.Abs(sourceDir)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get absolute path to source directory: %v", err)
			return err
		}

		prof.SetSourceDir(sourceDir)
	}

	for k, v := range parseEnvVars(setVars) {
		ov := prof.Get(k)
		ov.SetValue(v.Value())
		prof.Set(k, ov)
	}

	for _, v := range unsetVars {
		prof.Delete(v)
	}

	if newName != "" {
		err := profs.Add(newName, prof)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot rename profile %q -> %q: %v", name, newName, err)
			return err
		}

		err = profs.Delete(name)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot delete profile %q: %v", name, err)
			return err
		}
	} else {
		err := profs.Update(name, prof)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot update profile: %v", err)
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

func parseEnvVars(envVars []string) map[string]*profile.Var {
	env := make(map[string]*profile.Var, len(envVars))
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		env[strings.TrimSpace(parts[0])] = profile.NewVar(parts[1])
	}
	return env
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
