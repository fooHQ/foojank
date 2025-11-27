package edit

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
		Name:  "edit",
		Usage: "Edit configuration",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  flags.Set,
				Usage: "set configuration option (format: key=value)",
			},
			&cli.StringSliceFlag{
				Name:  flags.Unset,
				Usage: "unset configuration option (format: key)",
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
	setOptions, _ := conf.StringSlice(flags.Set)
	unsetOptions, _ := conf.StringSlice(flags.Unset)
	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	format, _ := conf.String(flags.Format)
	noColor, _ := conf.String(flags.NoColor)

	opts := map[string]string{
		config.FlagToOption(flags.ServerURL):         serverURL,
		config.FlagToOption(flags.ServerCertificate): serverCert,
		config.FlagToOption(flags.Account):           accountName,
		config.FlagToOption(flags.Format):            format,
		config.FlagToOption(flags.NoColor):           noColor,
	}
	for k, v := range config.ParseKVPairs(setOptions) {
		key := config.FlagToOption(k)
		_, ok := opts[key]
		if !ok {
			logger.ErrorContext(ctx, "Cannot set option %s: option not found", k)
			return errors.New("option not found")
		}
		opts[key] = v
	}
	for _, k := range unsetOptions {
		key := config.FlagToOption(k)
		delete(opts, key)
	}

	err := configdir.UpdateConfigJson(configDir, config.NewWithOptions(opts))
	if err != nil {
		logger.ErrorContext(ctx, "Cannot update configuration: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
