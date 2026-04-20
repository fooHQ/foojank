package config

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/commands/config/describe"
	"github.com/foohq/foojank/internal/commands/config/edit"
	confinit "github.com/foohq/foojank/internal/commands/config/init"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage configuration",
		Commands: []*cli.Command{
			confinit.NewCommand(),
			describe.NewCommand(),
			edit.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}
