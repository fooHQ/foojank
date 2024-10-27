package repository

import (
	"github.com/foojank/foojank/internal/application/commands/repository/copy"
	"github.com/foojank/foojank/internal/application/commands/repository/create"
	"github.com/foojank/foojank/internal/application/commands/repository/list"
	"github.com/urfave/cli/v2"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "repository",
		Description: "Manage file repositories.",
		Subcommands: []*cli.Command{
			create.NewCommand(),
			list.NewCommand(),
			copy.NewCommand(),
		},
		HideHelpCommand: true,
	}
}
