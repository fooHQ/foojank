package credential

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/commands/credential/describe"

	"github.com/foohq/foojank/cmd/foojank/commands/credential/create"
	"github.com/foohq/foojank/cmd/foojank/commands/credential/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "credential",
		Usage: "Manage credentials",
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
