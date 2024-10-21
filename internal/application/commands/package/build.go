package _package

import (
	"fmt"
	"github.com/foojank/foojank/fzz"
	"github.com/urfave/cli/v2"
	"path/filepath"
)

func NewBuildCommand() *cli.Command {
	return &cli.Command{
		Name:        "build",
		Description: "Build a package",
		Action:      newBuildCommandAction(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "name",
			},
		},
	}
}

func newBuildCommandAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		name := c.String("name")
		src := c.Args().Get(0)

		if name == "" {
			name = filepath.Base(src)
		}

		if src == "" {
			return fmt.Errorf("command expects an argument")
		}

		dst := fzz.NewFilename(name)
		err := fzz.Build(src, dst)
		if err != nil {
			return err
		}

		return nil
	}
}
