package job

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"

	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/job/create"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/job/describe"
	"github.com/foohq/foojank/cmd/foojank/commands/daemon/call/job/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "job",
		Usage: "Manage jobs",
		Commands: []*cli.Command{
			create.NewCommand(),
			describe.NewCommand(),
			list.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
