package set

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/configdir"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "set",
		Usage: "Set configuration option",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  config.ServerURL,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:  config.ServerCertificate,
				Usage: "set server TLS certificate",
			},
			&cli.StringFlag{
				Name:  config.Account,
				Usage: "set server account",
			},
			&cli.StringFlag{
				Name:  config.Format,
				Usage: "set output format",
			},
			&cli.BoolFlag{
				Name:  config.NoColor,
				Usage: "set disable color output",
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

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	configDir, _ := conf.String(config.ConfigDir)

	// Set ConfigDir to an empty value.
	conf = config.Merge(conf, config.NewWithOptions(map[string]any{
		config.ConfigDir: "",
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
