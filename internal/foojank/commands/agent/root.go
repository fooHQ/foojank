package agent

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/agent/build"
	"github.com/foohq/foojank/internal/foojank/commands/agent/list"
	"github.com/foohq/foojank/internal/foojank/commands/agent/logs"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "agent",
		Usage: "Manage agents",
		Commands: []*cli.Command{
			build.NewCommand(),
			list.NewCommand(),
			logs.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
