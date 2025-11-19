package profile

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/profile/create"
	"github.com/foohq/foojank/internal/foojank/commands/profile/edit"
	"github.com/foohq/foojank/internal/foojank/commands/profile/list"
	"github.com/foohq/foojank/internal/foojank/commands/profile/remove"
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
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
