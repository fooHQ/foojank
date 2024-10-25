package repository

import (
	"github.com/foojank/foojank/clients/repository"
	"github.com/urfave/cli/v2"
)

func NewRootCommand(repo *repository.Client) *cli.Command {
	return &cli.Command{
		Name:        "repository",
		Description: "Manage file repositories.",
		Subcommands: []*cli.Command{
			NewCreateCommand(repo),
			NewListCommand(repo),
		},
		HideHelpCommand: true,
	}
}
