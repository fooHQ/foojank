package application

import (
	"fmt"
	repoCli "github.com/foojank/foojank/clients/repository"
	"github.com/foojank/foojank/clients/vessel"
	"github.com/foojank/foojank/internal/application/commands/agent"
	_package "github.com/foojank/foojank/internal/application/commands/package"
	"github.com/foojank/foojank/internal/application/commands/repository"
	"github.com/urfave/cli/v2"
	"log/slog"
	"os"
)

type Arguments struct {
	Logger     *slog.Logger
	Vessel     *vessel.Client
	Repository *repoCli.Client
}

func New(args Arguments) *cli.App {
	return &cli.App{
		Name:     "foojank",
		HelpName: "foojank",
		Usage:    "Manage and control foojank agents",
		Args:     true,
		Version:  "0.1.0", // TODO: from config!
		Commands: []*cli.Command{
			agent.NewRootCommand(agent.Arguments{
				Logger: args.Logger,
				Vessel: args.Vessel,
			}),
			_package.NewRootCommand(_package.Arguments{
				Logger: args.Logger,
			}),
			repository.NewRootCommand(repository.Arguments{
				Logger:     args.Logger,
				Repository: args.Repository,
			}),
		},
		CommandNotFound: func(c *cli.Context, s string) {
			msg := fmt.Sprintf("command '%s %s' does not exist", c.Command.Name, s)
			args.Logger.Error(msg)
			os.Exit(1)
		},
		HideHelpCommand: true,
	}
}
