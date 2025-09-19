package repository

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/repository/copy"
	"github.com/foohq/foojank/internal/foojank/commands/repository/list"
	"github.com/foohq/foojank/internal/foojank/commands/repository/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "storage",
		Usage: "Manage storages",
		Commands: []*cli.Command{
			list.NewCommand(),
			copy.NewCommand(),
			remove.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
