package commands

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/cmd/foojankd/actions"
	"github.com/foohq/foojank/cmd/foojankd/commands/account"
	"github.com/foohq/foojank/cmd/foojankd/commands/role"
	"github.com/foohq/foojank/cmd/foojankd/commands/start"
	"github.com/foohq/foojank/cmd/foojankd/commands/user"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "foojankd",
		Usage:   "Daemon for Foojank C2 framework",
		Version: foojank.Version(),
		Commands: []*cli.Command{
			account.NewCommand(),
			role.NewCommand(),
			user.NewCommand(),
			start.NewCommand(),
		},
		CommandNotFound:       actions.CommandNotFound,
		OnUsageError:          actions.UsageError,
		HideHelpCommand:       true,
		EnableShellCompletion: true,
	}
}
