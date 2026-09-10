package daemon

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "daemon",
		Usage: "Daemon RPCs",
		Commands: []*cli.Command{
			call.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
