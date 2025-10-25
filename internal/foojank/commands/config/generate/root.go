package generate

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:            "generate",
		Usage:           "Generate configuration",
		Commands:        []*cli.Command{},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
