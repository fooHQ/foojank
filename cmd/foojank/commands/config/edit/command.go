package edit

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/configdir"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "edit",
		Usage: "Edit configuration",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  flags.Variable,
				Usage: "set configuration option (format: key=value)",
			},
			&cli.StringSliceFlag{
				Name:  flags.WithoutVariable,
				Usage: "unset configuration option (format: key)",
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

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, _ *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	configDir, _ := conf.String(flags.ConfigDir)
	setVars, _ := conf.StringSlice(flags.Variable)
	unsetVars, _ := conf.StringSlice(flags.WithoutVariable)
	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	format, _ := conf.String(flags.Format)
	noColor, _ := conf.String(flags.NoColor)

	if len(setVars) == 0 && len(unsetVars) == 0 {
		logger.ErrorContext(ctx, "Nothing to do.")
		return errors.New("nothing to do")
	}

	opts := map[string]string{
		flags.ServerURL:         serverURL,
		flags.ServerCertificate: serverCert,
		flags.Account:           accountName,
		flags.Format:            format,
		flags.NoColor:           noColor,
	}
	for k, v := range config.ParseKVPairs(setVars) {
		_, ok := opts[k]
		if !ok {
			logger.ErrorContext(ctx, "Cannot set option %s: option not found", k)
			return errors.New("option not found")
		}
		opts[k] = v
	}
	for _, k := range unsetVars {
		delete(opts, k)
	}

	err := configdir.UpdateConfigJSON(configDir, config.NewWithOptions(opts))
	if err != nil {
		logger.ErrorContext(ctx, "Cannot update configuration: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(_ *config.Config) error {
	return nil
}
