package user

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/commands/user/create"
	"github.com/foohq/foojank/cmd/foojankd/commands/user/describe"
	"github.com/foohq/foojank/cmd/foojankd/commands/user/list"

	"github.com/foohq/foojank/cmd/foojankd/actions"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "user",
		Usage: "Manage users",
		Commands: []*cli.Command{
			create.NewCommand(),
			describe.NewCommand(),
			list.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
