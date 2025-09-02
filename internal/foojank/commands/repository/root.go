package repository

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/repository/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "repository",
		Usage: "Manage repositories",
		Commands: []*cli.Command{
			list.NewCommand(),
			/*copy.NewCommand(),
			remove.NewCommand(),*/
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
