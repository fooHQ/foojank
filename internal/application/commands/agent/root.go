package agent

import (
	"github.com/foojank/foojank/internal/application/commands/agent/exec"
	"github.com/foojank/foojank/internal/application/commands/agent/list"
	"github.com/urfave/cli/v2"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "agent",
		Description: "Command & control installed agents.",
		Subcommands: []*cli.Command{
			list.NewCommand(),
			exec.NewCommand(),
		},
		HideHelpCommand: true,
	}
}
