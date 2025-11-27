package config

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/config/edit"
	confinit "github.com/foohq/foojank/internal/foojank/commands/config/init"
	"github.com/foohq/foojank/internal/foojank/commands/config/list"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage configuration",
		Commands: []*cli.Command{
			confinit.NewCommand(),
			edit.NewCommand(),
			list.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
