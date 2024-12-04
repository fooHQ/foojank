package _package

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/package/build"
	"github.com/foohq/foojank/internal/client/commands/package/inspect"
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
