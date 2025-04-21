package config

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/config/generate"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage configuration files",
		Commands: []*cli.Command{
			generate.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
	}
}
