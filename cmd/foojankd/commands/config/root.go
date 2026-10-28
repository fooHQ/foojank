package config

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/actions"
	"github.com/foohq/foojank/cmd/foojankd/commands/config/generate"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage server configuration",
		Commands: []*cli.Command{
			generate.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
