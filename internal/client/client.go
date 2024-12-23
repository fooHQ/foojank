package client

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/agent"
	_config "github.com/foohq/foojank/internal/client/commands/config"
	"github.com/foohq/foojank/internal/client/commands/repository"
	"github.com/foohq/foojank/internal/client/commands/script"
	"github.com/foohq/foojank/internal/client/flags"
	"github.com/foohq/foojank/internal/config"
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
				Value:   config.DefaultClientConfigPath(),
				Aliases: []string{"c"},
			},
			&cli.StringSliceFlag{
				Name:    flags.Server,
				Usage:   "set server URL",
				Value:   config.DefaultServers,
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:  flags.UserJWT,
				Usage: "set user JWT token",
			},
			&cli.StringFlag{
				Name:  flags.UserKey,
				Usage: "set user secret key",
			},
			&cli.StringFlag{
				Name:  flags.AccountJWT,
				Usage: "set account JWT token",
			},
			&cli.StringFlag{
				Name:  flags.AccountSigningKey,
				Usage: "set account signing key",
			},
			&cli.StringFlag{
				Name:  flags.Codebase,
				Usage: "set path to a directory with framework's source code",
			},
			&cli.StringFlag{
				Name:  flags.LogLevel,
				Usage: "set log level",
				Value: config.DefaultLogLevel,
			},
			&cli.BoolFlag{
				Name:  flags.NoColor,
				Usage: "disable color output",
				Value: config.DefaultNoColor,
			},
		},
		Commands: []*cli.Command{
			agent.NewCommand(),
			script.NewCommand(),
			repository.NewCommand(),
			_config.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
