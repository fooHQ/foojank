package profile

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/commands/profile/create"
	"github.com/foohq/foojank/internal/commands/profile/edit"
	_import "github.com/foohq/foojank/internal/commands/profile/import"
	"github.com/foohq/foojank/internal/commands/profile/list"
	"github.com/foohq/foojank/internal/commands/profile/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "profile",
		Usage: "Manage profiles",
		Commands: []*cli.Command{
			create.NewCommand(),
			edit.NewCommand(),
			list.NewCommand(),
			remove.NewCommand(),
			_import.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
