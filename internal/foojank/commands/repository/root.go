package repository

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/repository/copy"
	"github.com/foohq/foojank/internal/foojank/commands/repository/create"
	"github.com/foohq/foojank/internal/foojank/commands/repository/destroy"
	"github.com/foohq/foojank/internal/foojank/commands/repository/list"
	"github.com/foohq/foojank/internal/foojank/commands/repository/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "repository",
		Usage: "Manage repositories",
		Commands: []*cli.Command{
			create.NewCommand(),
			destroy.NewCommand(),
			list.NewCommand(),
			copy.NewCommand(),
			remove.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
