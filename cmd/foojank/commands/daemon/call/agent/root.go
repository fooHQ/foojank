package agent

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/agent/check"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/agent/describe"

	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/agent/build"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/agent/create"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/agent/list"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/agent/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "agent",
		Usage: "Manage agents",
		Commands: []*cli.Command{
			create.NewCommand(),
			describe.NewCommand(),
			build.NewCommand(),
			list.NewCommand(),
			check.NewCommand(),
			remove.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
