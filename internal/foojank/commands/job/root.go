package job

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/job/cancel"
	"github.com/foohq/foojank/internal/foojank/commands/job/create"
	"github.com/foohq/foojank/internal/foojank/commands/job/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "job",
		Usage: "Manage jobs",
		Commands: []*cli.Command{
			create.NewCommand(),
			cancel.NewCommand(),
			list.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
