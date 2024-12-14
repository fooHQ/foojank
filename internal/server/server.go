package server

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/server/actions"
	"github.com/foohq/foojank/internal/server/commands/start"
	"github.com/foohq/foojank/internal/server/flags"
)

var DefaultConfigFilename = "server.conf"

func New() *cli.Command {
	confDir, err := os.UserConfigDir()
	if err != nil {
		confDir = "./"
	}
	confPath := filepath.Join(confDir, "foojank", DefaultConfigFilename)

	return &cli.Command{
		Name:    "foojank",
		Usage:   "A foojank server",
		Version: foojank.Version(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flags.Config,
				Usage:   "path to a configuration file",
				Value:   confPath,
				Aliases: []string{"c"},
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
			start.NewCommand(),
		},
		DefaultCommand:  "start",
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
