package list

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config/v2"
	"github.com/foohq/foojank/internal/log"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Usage:   "List all scripts",
		Action:  action,
		Aliases: []string{"ls"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewClientConfig(ctx, c)
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
	// TODO: this should probably be defined in the config!
	codebaseDir := filepath.Join(*conf.DataDir, "src")
	client := codebase.New(codebaseDir)
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
