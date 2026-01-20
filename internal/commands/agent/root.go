package agent

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/commands/agent/build"
	"github.com/foohq/foojank/internal/commands/agent/list"
	"github.com/foohq/foojank/internal/commands/agent/logs"
	"github.com/foohq/foojank/internal/commands/agent/remove"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "agent",
		Usage: "Manage agents",
		Commands: []*cli.Command{
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
