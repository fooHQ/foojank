package foojank

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/account"
	"github.com/foohq/foojank/internal/foojank/commands/agent"
	"github.com/foohq/foojank/internal/foojank/commands/config"
	"github.com/foohq/foojank/internal/foojank/commands/job"
	"github.com/foohq/foojank/internal/foojank/commands/script"
	"github.com/foohq/foojank/internal/foojank/commands/server"
	"github.com/foohq/foojank/internal/foojank/commands/storage"
	"github.com/foohq/foojank/internal/foojank/flags"
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
			account.NewCommand(),
			agent.NewCommand(),
			job.NewCommand(),
			script.NewCommand(),
			storage.NewCommand(),
			config.NewCommand(),
			server.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
