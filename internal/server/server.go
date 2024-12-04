package server

// TODO: do not import from application directory!

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/server/actions"
	"github.com/foohq/foojank/internal/server/commands/start"
	"github.com/foohq/foojank/internal/server/flags"
)

func New() *cli.Command {
	return &cli.Command{
		Name:    "foojank",
		Usage:   "A foojank server",
		Version: foojank.Version(),
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  flags.LogLevel,
				Usage: "set log level",
				Value: 0,
			},
			&cli.BoolFlag{
				Name:  flags.NoColor,
				Usage: "disable color output",
			},
		},
		Commands: []*cli.Command{
			start.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
