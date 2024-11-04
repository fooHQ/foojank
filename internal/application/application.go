package application

import (
	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/application/actions"
	"github.com/foohq/foojank/internal/application/commands/agent"
	_package "github.com/foohq/foojank/internal/application/commands/package"
	"github.com/foohq/foojank/internal/application/commands/repository"
	"github.com/foohq/foojank/internal/application/flags"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:    "foojank",
		Usage:   "A cross-platform command and control (C2) framework",
		Version: foojank.Version(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flags.Server,
				Usage:   "URL of a NATS server",
				Value:   "wss://localhost",
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:    flags.Username,
				Usage:   "authenticate to the server as user",
				Aliases: []string{"u"},
			},
			&cli.StringFlag{
				Name:    flags.Password,
				Usage:   "set user password",
				Aliases: []string{"p"},
			},
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
			agent.NewCommand(),
			_package.NewCommand(),
			repository.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
