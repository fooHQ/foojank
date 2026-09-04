package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/commands/account"
	"github.com/foohq/foojank/cmd/foojank/commands/agent"
	"github.com/foohq/foojank/cmd/foojank/commands/config"
	"github.com/foohq/foojank/cmd/foojank/commands/gateway"
	"github.com/foohq/foojank/cmd/foojank/commands/job"
	"github.com/foohq/foojank/cmd/foojank/commands/profile"
	"github.com/foohq/foojank/cmd/foojank/commands/storage"
	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/clients/server"
	config2 "github.com/foohq/foojank/internal/config"
)

var app = &cli.Command{
	Name:    "foojank",
	Usage:   "Command and control framework",
	Version: foojank.Version(),
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  flags.NoColor,
			Usage: "disable color output",
		},
	},
	Commands: []*cli.Command{
		account.NewCommand(),
		agent.NewCommand(),
		config.NewCommand(),
		job.NewCommand(),
		profile.NewCommand(),
		storage.NewCommand(),
		gateway.NewCommand(),
		{
			Name: "init",
			Action: func(ctx context.Context, c *cli.Command) error {
				ctx, err := actions.LoadConfig(os.Stderr, func(conf *config2.Config) error {
					return nil
				})(ctx, c)
				if err != nil {
					return err
				}

				ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
				if err != nil {
					return err
				}

				conf := actions.GetConfigFromContext(ctx)
				logger := actions.GetLoggerFromContext(ctx)

				serverURL, _ := conf.String(flags.ServerURL)
				serverCert, _ := conf.String(flags.ServerCertificate)
				accountName, _ := conf.String(flags.Account)

				userJWT, userSeed, err := auth.ReadUser(accountName)
				if err != nil {
					logger.ErrorContext(ctx, "Cannot read user %q: %v", accountName, err)
					return err
				}

				srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
				if err != nil {
					logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
					return err
				}

				client := daemon.New(srv)
				err = client.InitDaemon(ctx)
				if err != nil {
					logger.ErrorContext(ctx, "Cannot initialize daemon: %v", err)
					return err
				}

				return nil
			},
		},
	},
	CommandNotFound: actions.CommandNotFound,
	OnUsageError:    actions.UsageError,
	HideHelpCommand: true,
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	err := app.Run(ctx, os.Args)
	if err != nil {
		cancel()
		os.Exit(1)
	}
	cancel()
}
