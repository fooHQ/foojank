package list

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/internal/client/actions"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List all scripts",
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)

	if conf.Codebase == nil {
		err := fmt.Errorf("cannot list scripts: codebase not configured")
		logger.Error(err.Error())
		return err
	}

	client := codebase.New(*conf.Codebase)
	return listAction(logger, client)(ctx, c)
}

func listAction(logger *slog.Logger, client *codebase.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		scripts, err := client.ListScripts()
		if err != nil {
			err := fmt.Errorf("cannot list scripts: codebase not configured")
			logger.Error(err.Error())
			return err
		}

		fmt.Printf("%+v\n", scripts)

		return nil
	}
}
