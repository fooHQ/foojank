package script

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/script/exec"
	"github.com/foohq/foojank/internal/client/commands/script/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "script",
		Usage: "Manage scripts",
		Commands: []*cli.Command{
			list.NewCommand(),
			exec.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
