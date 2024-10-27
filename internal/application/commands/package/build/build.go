package build

import (
	"fmt"
	"github.com/foojank/foojank/fzz"
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/urfave/cli/v2"
	"log/slog"
	"path/filepath"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "build",
		Description: "Build a package",
		Args:        true,
		ArgsUsage:   "<dir>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "name",
			},
		},
		Action: action,
	}
}

func action(c *cli.Context) error {
	logger := actions.NewLogger(c)
	return buildAction(logger)(c)
}

func buildAction(logger *slog.Logger) cli.ActionFunc {
	return func(c *cli.Context) error {
		cnt := c.Args().Len()
		if cnt != 1 {
			err := fmt.Errorf("command '%s' expects the following arguments: %s", c.Command.Name, c.Command.ArgsUsage)
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
