package user

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/user/create"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "user",
		Usage: "Manage users",
		Commands: []*cli.Command{
			create.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
