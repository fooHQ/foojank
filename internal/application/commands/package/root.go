package _package

import (
	"github.com/urfave/cli/v2"
)

func NewRootCommand() *cli.Command {
	return &cli.Command{
		Name:        "package",
		Description: "Manage fzz packages.",
		Subcommands: []*cli.Command{
			NewBuildCommand(),
		},
		HideHelpCommand: true,
	}
}
