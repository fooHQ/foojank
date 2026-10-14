package commands

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/commands/agent"
	"github.com/foohq/foojank/cmd/foojank/commands/config"
	"github.com/foohq/foojank/cmd/foojank/commands/credential"
	"github.com/foohq/foojank/cmd/foojank/commands/gateway"
	"github.com/foohq/foojank/cmd/foojank/commands/job"
	"github.com/foohq/foojank/cmd/foojank/commands/profile"
	"github.com/foohq/foojank/cmd/foojank/commands/role"
	"github.com/foohq/foojank/cmd/foojank/commands/storage"
	"github.com/foohq/foojank/cmd/foojank/commands/user"
	"github.com/foohq/foojank/cmd/foojank/flags"
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
			agent.NewCommand(),
			config.NewCommand(),
			credential.NewCommand(),
			gateway.NewCommand(),
			job.NewCommand(),
			profile.NewCommand(),
			role.NewCommand(),
			storage.NewCommand(),
			user.NewCommand(),
		},
		CommandNotFound:       actions.CommandNotFound,
		OnUsageError:          actions.UsageError,
		HideHelpCommand:       true,
		EnableShellCompletion: true,
	}
}
