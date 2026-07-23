package storage

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"

	"github.com/foohq/foojank/cmd/foojank/commands/storage/copy"
	"github.com/foohq/foojank/cmd/foojank/commands/storage/list"
	"github.com/foohq/foojank/cmd/foojank/commands/storage/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "storage",
		Usage: "Manage storage",
		Commands: []*cli.Command{
			list.NewCommand(),
			copy.NewCommand(),
			remove.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
