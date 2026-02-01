package job

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/commands/job/attach"
	"github.com/foohq/foojank/internal/commands/job/cancel"
	"github.com/foohq/foojank/internal/commands/job/create"
	"github.com/foohq/foojank/internal/commands/job/list"
	"github.com/foohq/foojank/internal/commands/job/logs"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "job",
		Usage: "Manage jobs",
		Commands: []*cli.Command{
			create.NewCommand(),
			cancel.NewCommand(),
			list.NewCommand(),
			logs.NewCommand(),
			attach.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
