package job

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"

	"github.com/foohq/foojank/cmd/foojank/commands/job/cancel"
	"github.com/foohq/foojank/cmd/foojank/commands/job/create"
	"github.com/foohq/foojank/cmd/foojank/commands/job/describe"
	"github.com/foohq/foojank/cmd/foojank/commands/job/list"
	"github.com/foohq/foojank/cmd/foojank/commands/job/logs"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "job",
		Usage: "Manage jobs",
		Commands: []*cli.Command{
			create.NewCommand(),
			describe.NewCommand(),
			cancel.NewCommand(),
			list.NewCommand(),
			logs.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
