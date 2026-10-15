package user

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/commands/user/create"
	"github.com/foohq/foojank/cmd/foojank/commands/user/describe"
	"github.com/foohq/foojank/cmd/foojank/commands/user/edit"
	"github.com/foohq/foojank/cmd/foojank/commands/user/list"
	"github.com/foohq/foojank/cmd/foojank/commands/user/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "user",
		Usage: "Manage users",
		Commands: []*cli.Command{
			create.NewCommand(),
			describe.NewCommand(),
			edit.NewCommand(),
			list.NewCommand(),
			remove.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
