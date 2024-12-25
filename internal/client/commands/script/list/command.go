package list

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config"
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
	conf, err := actions.NewConfig(ctx, c, validateConfiguration)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)
	client := codebase.New(*conf.Codebase)
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
			_, _ = fmt.Fprint(os.Stdout, script)
		}

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.Codebase == nil {
		return errors.New("codebase not configured")
	}

	return nil
}
