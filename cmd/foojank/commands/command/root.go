package command

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"

	"github.com/foohq/foojank/cmd/foojank/commands/command/describe"
	"github.com/foohq/foojank/cmd/foojank/commands/command/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "command",
		Usage: "Manage commands",
		Commands: []*cli.Command{
			describe.NewCommand(),
			list.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
