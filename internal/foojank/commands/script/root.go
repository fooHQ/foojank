package script

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/script/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "script",
		Usage: "Manage scripts",
		Commands: []*cli.Command{
			list.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
