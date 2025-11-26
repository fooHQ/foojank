package set

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/configdir"
	"github.com/foohq/foojank/internal/foojank/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "set",
		Usage: "Set configuration option",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.ServerURL,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:      flags.ServerCertificate,
				Usage:     "set path to server's certificate",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:  flags.Account,
				Usage: "set server account",
			},
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
			},
			&cli.BoolFlag{
				Name:  flags.NoColor,
				Usage: "set disable color output",
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

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	configDir, _ := conf.String(flags.ConfigDir)

	// Set ConfigDir to an empty value.
	conf = config.Merge(conf, config.NewWithOptions(map[string]any{
		flags.ConfigDir: "",
	}))

	err := configdir.UpdateConfigJson(configDir, conf)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot update configuration: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
