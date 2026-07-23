package agent

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"

	"github.com/foohq/foojank/cmd/foojank/commands/agent/build"
	"github.com/foohq/foojank/cmd/foojank/commands/agent/create"
	"github.com/foohq/foojank/cmd/foojank/commands/agent/list"
	"github.com/foohq/foojank/cmd/foojank/commands/agent/logs"
	"github.com/foohq/foojank/cmd/foojank/commands/agent/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "agent",
		Usage: "Manage agents",
		Commands: []*cli.Command{
			create.NewCommand(),
			build.NewCommand(),
			list.NewCommand(),
			remove.NewCommand(),
			logs.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
