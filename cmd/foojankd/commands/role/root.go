package role

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/commands/role/create"
	"github.com/foohq/foojank/cmd/foojankd/commands/role/describe"
	importcmd "github.com/foohq/foojank/cmd/foojankd/commands/role/import"
	"github.com/foohq/foojank/cmd/foojankd/commands/role/list"
	"github.com/foohq/foojank/cmd/foojankd/commands/role/remove"

	"github.com/foohq/foojank/cmd/foojankd/actions"
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
			importcmd.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
