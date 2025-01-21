package server

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/server/start"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "Manage server",
		Commands: []*cli.Command{
			start.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
