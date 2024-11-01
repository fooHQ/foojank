package _package

import (
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/foojank/foojank/internal/application/commands/package/build"
	"github.com/foojank/foojank/internal/application/commands/package/inspect"
	"github.com/urfave/cli/v3"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "package",
		Usage: "Manage packages",
		Commands: []*cli.Command{
			build.NewCommand(),
			inspect.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
