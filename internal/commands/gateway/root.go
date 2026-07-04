package gateway

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/commands/gateway/create"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "gateway",
		Usage: "Manage gateway configurations",
		Commands: []*cli.Command{
			create.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
