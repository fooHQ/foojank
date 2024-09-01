package application

import (
	"github.com/foojank/foojank/internal/application/commands"
	"github.com/nats-io/nats.go"
	"github.com/urfave/cli/v2"
)

func New(nc *nats.Conn) *cli.App {
	return &cli.App{
		Name:        "foojank",
		HelpName:    "foojank",
		Usage:       "BLABLA 1",
		UsageText:   "BLABLA 2",
		Args:        false,
		ArgsUsage:   "AAAA",
		Version:     "0.1.0", // TODO: from config!
		Description: "DDDDD",
		Commands: []*cli.Command{
			commands.NewListCommand(nc),
			commands.NewRunCommand(nc),
			commands.NewExitCommand(),
		},
		CommandNotFound: func(c *cli.Context, s string) {
			// TODO: refactor!
			println("unknown command")
		},
	}
}
