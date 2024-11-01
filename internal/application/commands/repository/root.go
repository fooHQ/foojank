package repository

import (
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/foojank/foojank/internal/application/commands/repository/copy"
	"github.com/foojank/foojank/internal/application/commands/repository/create"
	"github.com/foojank/foojank/internal/application/commands/repository/destroy"
	"github.com/foojank/foojank/internal/application/commands/repository/list"
	"github.com/urfave/cli/v3"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "repository",
		Description: "Manage file repositories.",
		Commands: []*cli.Command{
			create.NewCommand(),
			destroy.NewCommand(),
			list.NewCommand(),
			copy.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
