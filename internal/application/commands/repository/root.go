package repository

import (
	"github.com/foojank/foojank/clients/repository"
	"github.com/urfave/cli/v2"
	"log/slog"
)

type Arguments struct {
	Logger     *slog.Logger
	Repository *repository.Client
}

func NewRootCommand(args Arguments) *cli.Command {
	return &cli.Command{
		Name:        "repository",
		Description: "Manage file repositories.",
		Subcommands: []*cli.Command{
			NewCreateCommand(CreateArguments{
				Logger:     args.Logger,
				Repository: args.Repository,
			}),
			NewListCommand(ListArguments{
				Logger:     args.Logger,
				Repository: args.Repository,
			}),
		},
		HideHelpCommand: true,
	}
}
