package application

import (
	"fmt"
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/foojank/foojank/internal/application/commands/agent"
	_package "github.com/foojank/foojank/internal/application/commands/package"
	"github.com/foojank/foojank/internal/application/commands/repository"
	"github.com/foojank/foojank/internal/application/flags"
	"github.com/urfave/cli/v2"
	"os"
)

func New() *cli.App {
	return &cli.App{
		Name:     "foojank",
		HelpName: "foojank",
		Usage:    "A cross-platform command and control (C2) framework",
		Args:     true,
		Version:  "0.1.0", // TODO: from config!
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flags.Server,
				Usage:   "URL of a NATS server",
				Value:   "wss://localhost",
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:    flags.User,
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
		},
		Commands: []*cli.Command{
			agent.NewCommand(),
			_package.NewCommand(),
			repository.NewCommand(),
		},
		CommandNotFound: func(c *cli.Context, s string) {
			logger := actions.NewLogger(c)
			msg := fmt.Sprintf("command '%s %s' does not exist", c.Command.Name, s)
			logger.Error(msg)
			os.Exit(1)
		},
		HideHelpCommand: true,
	}
}
