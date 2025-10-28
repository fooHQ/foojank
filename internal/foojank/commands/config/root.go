package config

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	confinit "github.com/foohq/foojank/internal/foojank/commands/config/init"
	"github.com/foohq/foojank/internal/foojank/commands/config/list"
	"github.com/foohq/foojank/internal/foojank/commands/config/set"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage configuration",
		Commands: []*cli.Command{
			confinit.NewCommand(),
			set.NewCommand(),
			list.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		OnUsageError:    actions.UsageError,
	}
}
