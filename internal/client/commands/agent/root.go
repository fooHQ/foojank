package agent

import (
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/agent/exec"
	"github.com/foohq/foojank/internal/client/commands/agent/list"
	"github.com/urfave/cli/v3"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "agent",
		Usage: "Manage agents",
		Commands: []*cli.Command{
			list.NewCommand(),
			exec.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
