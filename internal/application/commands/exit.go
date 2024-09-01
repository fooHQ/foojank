package commands

import (
	"github.com/urfave/cli/v2"
)

func NewExitCommand() *cli.Command {
	return &cli.Command{
		Name:   "exit",
		Action: newExitCommandAction(),
	}
}

func newExitCommandAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		return nil
	}
}
