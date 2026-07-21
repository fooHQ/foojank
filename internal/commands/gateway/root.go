package gateway

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/commands/gateway/create"
	"github.com/foohq/foojank/internal/commands/gateway/describe"
	"github.com/foohq/foojank/internal/commands/gateway/list"
	"github.com/foohq/foojank/internal/commands/gateway/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "gateway",
		Usage: "Manage gateway configurations",
		Commands: []*cli.Command{
			create.NewCommand(),
			describe.NewCommand(),
			list.NewCommand(),
			remove.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
