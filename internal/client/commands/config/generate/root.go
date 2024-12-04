package generate

import (
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/commands/config/generate/client"
	"github.com/foohq/foojank/internal/client/commands/config/generate/seed"
	"github.com/foohq/foojank/internal/client/commands/config/generate/server"
	"github.com/urfave/cli/v3"
)

// Usage:
// # Generate commands print yaml to stdout
// config create > seed.yaml

// config generate --type=server <seed-file>

// # Generate manager configuration file.
// # This also creates a new USER nkey and JWT signed by a selected ACCOUNT.
// # This configuration is meant to be used with foojank client application.
// config generate --type=client <seed-file>

// # Generate agent configuration file.
// # This also creates a new USER nkey and JWT signed by a selected ACCOUNT.
// # This configuration is meant to be used with vessel agent.
// config generate --type=agent <seed-file>

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
			client.NewCommand(),
			server.NewCommand(),
			seed.NewCommand(),
		},
		CommandNotFound: actions.CommandNotFound,
		HideHelpCommand: true,
	}
}
