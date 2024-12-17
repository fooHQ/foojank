package client

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/agent"
	"github.com/foohq/foojank/internal/client/commands/config"
	_package "github.com/foohq/foojank/internal/client/commands/package"
	"github.com/foohq/foojank/internal/client/commands/repository"
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
				Usage:   "path to a configuration file",
				Value:   flags.DefaultConfig(),
				Aliases: []string{"c"},
			},
			// TODO: use string slice!
			&cli.StringFlag{
				Name:    flags.Server,
				Usage:   "server URL",
				Value:   flags.DefaultServer[0],
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:  flags.UserJWT,
				Usage: "user JWT token",
			},
			&cli.StringFlag{
				Name:  flags.UserKey,
				Usage: "user secret NKey",
			},
			&cli.StringFlag{
				Name:  flags.AccountJWT,
				Usage: "account JWT token",
			},
			&cli.StringFlag{
				Name:  flags.AccountSigningKey,
				Usage: "account signing key",
			},
			&cli.StringFlag{
				Name:  flags.Codebase,
				Usage: "path to directory with foojank codebase",
				Value: flags.DefaultCodebase(),
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
			agent.NewCommand(),
			_package.NewCommand(),
			repository.NewCommand(),
			config.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
