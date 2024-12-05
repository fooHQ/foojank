package generate

import (
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/config/generate/agent"
	"github.com/foohq/foojank/internal/client/commands/config/generate/client"
	"github.com/foohq/foojank/internal/client/commands/config/generate/seed"
	"github.com/foohq/foojank/internal/client/commands/config/generate/server"
)

// Usage:
// # Generate commands print yaml to stdout
// config generate seed

// config generate server <seed-file>

// # Generate manager configuration file.
// # This also creates a new USER nkey and JWT signed by a selected ACCOUNT.
// # This configuration is meant to be used with foojank client application.
// config generate client <file>

// # Generate agent configuration file.
// # This also creates a new USER nkey and JWT signed by a selected ACCOUNT.
// # This configuration is meant to be used with vessel agent.
// config generate agent

// # For debugging only.
// config generate --type=creds <seed-file>
// TODO: use https://pkg.go.dev/github.com/nats-io/jwt/v2#DecorateJWT

// # Generate and add to a seed file another account.
// # Requires reloading the server configuration.
// config account(s) add
// config account(s) delete
// config account(s) import

// # Add another operator (maybe in the future!).
// config operator(s) add
// config operator(s) import

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate configuration files",
		Commands: []*cli.Command{
			agent.NewCommand(),
			client.NewCommand(),
			server.NewCommand(),
			seed.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
