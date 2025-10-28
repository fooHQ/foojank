package set

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
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
				Name:  flags.ServerCertificate,
				Usage: "set server TLS certificate",
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
				Usage: "disable color output",
			},
		},
		Action:       action,
		OnUsageError: actions.UsageError,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	configDir, _ := conf.String(flags.ConfigDir)

	if len(c.LocalFlagNames()) == 0 {
		return nil
	}

	// Parse flags once again but this time parse only the local flags
	confFlags, err := config.ParseFlags(c.LocalFlagNames(), func(name string) (any, bool) {
		return c.Value(name), c.IsSet(name)
	})
	if err != nil {
		log.Error(ctx, "Cannot parse command options: %v", err)
		return err
	}

	conf = config.Merge(conf, confFlags)

	err = actions.UpdateConfigJson(configDir, conf)
	if err != nil {
		log.Error(ctx, "Cannot update config json: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
