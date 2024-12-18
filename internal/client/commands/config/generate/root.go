package generate

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/config/generate/client"
	"github.com/foohq/foojank/internal/client/commands/config/generate/master"
	"github.com/foohq/foojank/internal/client/commands/config/generate/server"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate configuration files",
		Commands: []*cli.Command{
			master.NewCommand(),
			client.NewCommand(),
			server.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
