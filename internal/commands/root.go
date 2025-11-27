package commands

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/commands/account"
	"github.com/foohq/foojank/internal/commands/agent"
	"github.com/foohq/foojank/internal/commands/config"
	"github.com/foohq/foojank/internal/commands/job"
	"github.com/foohq/foojank/internal/commands/profile"
	"github.com/foohq/foojank/internal/commands/storage"
	"github.com/foohq/foojank/internal/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "foojank",
		Usage:   "Command and control framework",
		Version: foojank.Version(),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  flags.NoColor,
				Usage: "disable color output",
			},
		},
		Commands: []*cli.Command{
			account.NewCommand(),
			agent.NewCommand(),
			config.NewCommand(),
			job.NewCommand(),
			profile.NewCommand(),
			storage.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
