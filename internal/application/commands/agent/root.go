package agent

import (
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/foojank/foojank/internal/application/commands/agent/exec"
	"github.com/foojank/foojank/internal/application/commands/agent/list"
	"github.com/urfave/cli/v3"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "agent",
		Description: "Command & control installed agents.",
		Commands: []*cli.Command{
			list.NewCommand(),
			exec.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
