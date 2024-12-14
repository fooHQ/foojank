package agent

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/agent/build"
	"github.com/foohq/foojank/internal/client/commands/agent/exec"
	"github.com/foohq/foojank/internal/client/commands/agent/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "agent",
		Usage: "Manage agents",
		Commands: []*cli.Command{
			build.NewCommand(),
			list.NewCommand(),
			exec.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
