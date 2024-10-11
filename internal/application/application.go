package application

import (
	"github.com/foojank/foojank/clients/vessel"
	"github.com/foojank/foojank/internal/application/commands"
	"github.com/urfave/cli/v2"
)

func New(vessel *vessel.Client) *cli.App {
	return &cli.App{
		Name:     "foojank",
		HelpName: "foojank",
		Usage:    "Command & Control connected agents",
		Args:     true,
		Version:  "0.1.0", // TODO: from config!
		Commands: []*cli.Command{
			commands.NewListCommand(vessel),
			commands.NewRunCommand(vessel),
		},
		CommandNotFound: func(c *cli.Context, s string) {
			// TODO: refactor!
			println("unknown command")
		},
	}
}
