package role

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"

	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/role/create"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/role/describe"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/role/list"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/role/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "role",
		Usage: "Manage roles",
		Commands: []*cli.Command{
			create.NewCommand(),
			describe.NewCommand(),
			list.NewCommand(),
			remove.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
