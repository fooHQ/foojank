package storage

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	copy2 "github.com/foohq/foojank/internal/commands/storage/copy"
	"github.com/foohq/foojank/internal/commands/storage/list"
	"github.com/foohq/foojank/internal/commands/storage/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "storage",
		Usage: "Manage storage",
		Commands: []*cli.Command{
			list.NewCommand(),
			copy2.NewCommand(),
			remove.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
