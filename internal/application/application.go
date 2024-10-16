package application

import (
	"github.com/foojank/foojank/clients/vessel"
	"github.com/foojank/foojank/internal/application/commands/agent"
	_package "github.com/foojank/foojank/internal/application/commands/package"
	"github.com/urfave/cli/v2"
)

func New(vessel *vessel.Client) *cli.App {
	return &cli.App{
		Name:     "foojank",
		HelpName: "foojank",
		Usage:    "Manage and control foojank agents",
		Args:     true,
		Version:  "0.1.0", // TODO: from config!
		Commands: []*cli.Command{
			agent.NewRootCommand(vessel),
			_package.NewRootCommand(),
		},
		CommandNotFound: func(c *cli.Context, s string) {
			// TODO: refactor!
			println("unknown command")
		},
		HideHelpCommand: true,
	}
}
