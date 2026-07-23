package profile

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"

	"github.com/foohq/foojank/cmd/foojank/commands/profile/create"
	"github.com/foohq/foojank/cmd/foojank/commands/profile/describe"
	"github.com/foohq/foojank/cmd/foojank/commands/profile/edit"
	_import "github.com/foohq/foojank/cmd/foojank/commands/profile/import"
	"github.com/foohq/foojank/cmd/foojank/commands/profile/list"
	"github.com/foohq/foojank/cmd/foojank/commands/profile/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "profile",
		Usage: "Manage profiles",
		Commands: []*cli.Command{
			create.NewCommand(),
			describe.NewCommand(),
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
