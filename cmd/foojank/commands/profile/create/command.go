package create

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
		Name:  "create",
		Usage: "Create profile",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Name,
				Usage: "set profile name",
			},
			&cli.StringFlag{
				Name:  flags.Os,
				Usage: "set OS variable",
			},
			&cli.StringFlag{
				Name:  flags.Arch,
				Usage: "set ARCH variable",
			},
			&cli.StringSliceFlag{
				Name:  flags.Variable,
				Usage: "set config variable (format: key=value)",
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

	name, _ := conf.String(flags.Name)
	configDir, _ := conf.String(flags.ConfigDir)
	targetOS, _ := conf.String(flags.Os)
	targetArch, _ := conf.String(flags.Arch)
	setVars, _ := conf.StringSlice(flags.Variable)

	prof := profile.NewProfile()
	if targetOS != "" {
		prof.SetOS(targetOS)
	}

	if targetArch != "" {
		prof.SetArch(targetArch)
	}

	for k, v := range profile.ParseKVPairs(setVars) {
		prof.Set(k, v)
	}

	err := profs.Add(name, prof)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create profile: %v", err)
		return err
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
		flags.Name,
	} {
		switch opt {
		case flags.Name:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("profile name not configured")
			}
		}
	}
	return nil
}
