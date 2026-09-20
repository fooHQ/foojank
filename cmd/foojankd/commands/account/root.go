package account

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/actions"

	"github.com/foohq/foojank/cmd/foojankd/commands/account/create"
	"github.com/foohq/foojank/cmd/foojankd/commands/account/describe"
	"github.com/foohq/foojank/cmd/foojankd/commands/account/edit"
	"github.com/foohq/foojank/cmd/foojankd/commands/account/export"
	"github.com/foohq/foojank/cmd/foojankd/commands/account/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "account",
		Usage: "Manage accounts",
		Commands: []*cli.Command{
			create.NewCommand(),
			describe.NewCommand(),
			edit.NewCommand(),
			export.NewCommand(),
			list.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
