package call

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/user"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "call",
		Usage: "Call a daemon procedure",
		Commands: []*cli.Command{
			user.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
