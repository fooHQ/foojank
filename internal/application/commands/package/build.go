package _package

import (
	"fmt"
	"github.com/foojank/foojank/fzz"
	"github.com/urfave/cli/v2"
	"log/slog"
	"path/filepath"
)

type BuildArguments struct {
	Logger *slog.Logger
}

func NewBuildCommand(args BuildArguments) *cli.Command {
	return &cli.Command{
		Name:        "build",
		Description: "Build a package",
		Args:        true,
		ArgsUsage:   "<dir>",
		Action:      newBuildCommandAction(args),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "name",
			},
		},
	}
}

func newBuildCommandAction(args BuildArguments) cli.ActionFunc {
	return func(c *cli.Context) error {
		cnt := c.Args().Len()
		if cnt != 1 {
			err := fmt.Errorf("command '%s' expects the following arguments: %s", c.Command.Name, c.Command.ArgsUsage)
			args.Logger.Error(err.Error())
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
			args.Logger.Error(err.Error())
			return err
		}

		return nil
	}
}
