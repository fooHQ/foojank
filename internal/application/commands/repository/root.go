package repository

import (
	"github.com/foohq/foojank/internal/application/actions"
	"github.com/foohq/foojank/internal/application/commands/repository/copy"
	"github.com/foohq/foojank/internal/application/commands/repository/create"
	"github.com/foohq/foojank/internal/application/commands/repository/destroy"
	"github.com/foohq/foojank/internal/application/commands/repository/list"
	"github.com/foohq/foojank/internal/application/commands/repository/remove"
	"github.com/urfave/cli/v3"
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
		HideHelpCommand: true,
	}
}
