package commands

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/cmd/foojankgw/commands/credential"

	"github.com/foohq/foojank/cmd/foojankgw/actions"

	"github.com/foohq/foojank/cmd/foojankgw/commands/check"
	"github.com/foohq/foojank/cmd/foojankgw/commands/start"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "foojank-http-gateway",
		Usage:   "HTTP Gateway for Foojank C2 framework",
		Version: foojank.Version(),
		Commands: []*cli.Command{
			credential.NewCommand(),
			check.NewCommand(),
			start.NewCommand(),
		},
		CommandNotFound:       actions.CommandNotFound,
		OnUsageError:          actions.UsageError,
		HideHelpCommand:       true,
		EnableShellCompletion: true,
	}
}
