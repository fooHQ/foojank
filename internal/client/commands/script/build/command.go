package build

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/fzz"
	"github.com/foohq/foojank/internal/client/actions"
)

const (
	FlagOutput = "output"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "build",
		ArgsUsage: "<script-name>",
		Usage:     "Build a package",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    FlagOutput,
				Usage:   "set output file",
				Aliases: []string{"o"},
			},
		},
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
		err := fmt.Errorf("cannot build a package: codebase not configured")
		logger.Error(err.Error())
		return err
	}

	client := codebase.New(*conf.Codebase)
	return buildAction(logger, client)(ctx, c)
}

func buildAction(logger *slog.Logger, client *codebase.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		scriptName := c.Args().First()
		scriptPath, err := client.GetScript(scriptName)
		if err != nil {
			if os.IsNotExist(err) {
				err := fmt.Errorf("cannot build a package: script '%s' not found", scriptName)
				logger.Error(err.Error())
				return err
			}

			err := fmt.Errorf("cannot build a package: %v", err)
			logger.Error(err.Error())
			return err
		}

		src := scriptPath
		name := c.String(FlagOutput)
		if name == "" {
			name = filepath.Base(src)
		}

		dst := fzz.NewFilename(name)
		err = fzz.Build(src, dst)
		if err != nil {
			err := fmt.Errorf("cannot build a package: %v", err)
			logger.Error(err.Error())
			return err
		}

		return nil
	}
}
