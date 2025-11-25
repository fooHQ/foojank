package create

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
	"github.com/foohq/foojank/internal/profile"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		ArgsUsage: "<name>",
		Usage:     "Create a new profile",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  config.SourceDir,
				Usage: "set path to a source code directory",
			},
			&cli.StringSliceFlag{
				Name:  config.Set,
				Usage: "set environment variable (format: KEY=value)",
			},
			&cli.StringFlag{
				Name:  config.ConfigDir,
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

	configDir, _ := conf.String(config.ConfigDir)
	sourceDir, _ := conf.String(config.SourceDir)
	setVars, _ := conf.StringSlice(config.Set)

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
	for k, v := range parseEnvVars(setVars) {
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
