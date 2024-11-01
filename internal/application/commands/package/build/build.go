package build

import (
	"context"
	"fmt"
	"github.com/foojank/foojank/fzz"
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/urfave/cli/v3"
	"log/slog"
	"path/filepath"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "build",
		Description: "Build a package",
		//Args:        true,
		ArgsUsage: "<dir>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "name",
			},
		},
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	logger := actions.NewLogger(ctx, c)

	if c.Args().Len() != 1 {
		err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
		logger.Error(err.Error())
		return err
	}

	return buildAction(logger)(ctx, c)
}

func buildAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
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
