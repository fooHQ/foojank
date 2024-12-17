package server

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
			&cli.StringFlag{
				Name:    flags.Config,
				Usage:   "path to a configuration file",
				Value:   flags.DefaultConfig(),
				Aliases: []string{"c"},
			},
			&cli.StringFlag{
				Name:  flags.Host,
				Usage: "bind to host",
				Value: flags.DefaultHost,
			},
			&cli.IntFlag{
				Name:  flags.Port,
				Usage: "bind to port",
				Value: flags.DefaultPort,
			},
			&cli.StringFlag{
				Name:     flags.OperatorJWT,
				Usage:    "operator JWT token",
				Required: true,
			},
			&cli.StringFlag{
				Name:     flags.SystemAccountJWT,
				Usage:    "system account JWT token",
				Required: true,
			},
			&cli.StringFlag{
				Name:     flags.AccountJWT,
				Usage:    "account JWT token",
				Required: true,
			},
			&cli.IntFlag{
				Name:  flags.LogLevel,
				Usage: "set log level",
				Value: flags.DefaultLogLevel,
			},
			&cli.BoolFlag{
				Name:  flags.NoColor,
				Usage: "disable color output",
				Value: flags.DefaultNoColor,
			},
		},
		Commands: []*cli.Command{
			start.NewCommand(),
		},
		DefaultCommand:  "start",
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
