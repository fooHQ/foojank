package list

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagDataDir = flags.DataDir
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all scripts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  FlagDataDir,
				Usage: "set path to a data directory",
			},
		},
		Action:  action,
		Aliases: []string{"ls"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: cannot parse configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)
	client := codebase.New(*conf.DataDir)
	return listAction(logger, client)(ctx, c)
}

func listAction(logger *slog.Logger, client *codebase.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		scripts, err := client.ListScripts()
		if err != nil {
			err := fmt.Errorf("cannot list scripts: codebase not configured")
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		for _, script := range scripts {
			_, _ = fmt.Fprintln(os.Stdout, script)
		}

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.LogLevel == nil {
		return errors.New("log level not configured")
	}

	if conf.NoColor == nil {
		return errors.New("no color not configured")
	}

	if conf.DataDir == nil {
		return errors.New("codebase not configured")
	}

	return nil
}
