package build

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/fzz"
	"github.com/foohq/foojank/internal/client/actions"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "build",
		ArgsUsage: "<dir>",
		Usage:     "Build a new package",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "set output name",
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

	return buildAction(logger)(ctx, c)
}

func buildAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		src := c.Args().Get(0)
		name := c.String("name")
		if name == "" {
			name = filepath.Base(src)
		}

		dst := fzz.NewFilename(name)
		err := fzz.Build(src, dst)
		if err != nil {
			err := fmt.Errorf("cannot build a package: %v", err)
			logger.Error(err.Error())
			return err
		}

		return nil
	}
}
