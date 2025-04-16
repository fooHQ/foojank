package generate

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/commands/config/generate/client"
	"github.com/foohq/foojank/internal/foojank/commands/config/generate/server"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate configuration",
		Commands: []*cli.Command{
			server.NewCommand(),
			client.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
