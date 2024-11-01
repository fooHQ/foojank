package _package

import (
	"github.com/foojank/foojank/internal/application/commands/package/build"
	"github.com/urfave/cli/v3"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "package",
		Description: "Manage fzz packages.",
		Commands: []*cli.Command{
			build.NewCommand(),
		},
		HideHelpCommand: true,
	}
}
