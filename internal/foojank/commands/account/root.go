package account

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/account/create"
	"github.com/foohq/foojank/internal/foojank/commands/account/export"
	"github.com/foohq/foojank/internal/foojank/commands/account/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "account",
		Usage: "Manage accounts",
		Commands: []*cli.Command{
			create.NewCommand(),
			export.NewCommand(),
			list.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
