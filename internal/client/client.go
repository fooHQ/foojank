package client

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/agent"
	"github.com/foohq/foojank/internal/client/commands/config"
	"github.com/foohq/foojank/internal/client/commands/repository"
	"github.com/foohq/foojank/internal/client/commands/script"
	"github.com/foohq/foojank/internal/client/commands/server"
	"github.com/foohq/foojank/internal/client/flags"
)

func New() *cli.Command {
	return &cli.Command{
		Name:    "foojank",
		Usage:   "A cross-platform command and control (C2) framework",
		Version: foojank.Version(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flags.Config,
				Usage:   "set path to a configuration file",
				Aliases: []string{"c"},
			},
			&cli.StringFlag{
				Name:  flags.LogLevel,
				Usage: "set log level",
			},
			&cli.BoolFlag{
				Name:  flags.NoColor,
				Usage: "disable color output",
			},
		},
		Commands: []*cli.Command{
			agent.NewCommand(),
			script.NewCommand(),
			repository.NewCommand(),
			config.NewCommand(),
			server.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
