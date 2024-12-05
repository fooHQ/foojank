package client

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/agent"
	"github.com/foohq/foojank/internal/client/commands/config"
	_package "github.com/foohq/foojank/internal/client/commands/package"
	"github.com/foohq/foojank/internal/client/commands/repository"
	"github.com/foohq/foojank/internal/client/flags"
)

const DefaultConfigFilename = "fjrc.toml"

func New() *cli.Command {
	confDir, err := os.UserConfigDir()
	if err != nil {
		confDir = "./"
	}
	confPath := filepath.Join(confDir, "foojank", DefaultConfigFilename)

	return &cli.Command{
		Name:    "foojank",
		Usage:   "A cross-platform command and control (C2) framework",
		Version: foojank.Version(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flags.Config,
				Usage:   "path to a configuration file",
				Value:   confPath,
				Aliases: []string{"c"},
			},
			// TODO: use string slice!
			&cli.StringFlag{
				Name:    flags.Server,
				Usage:   "server URL",
				Value:   "localhost",
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
			config.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
