package repository

import (
	"fmt"
	"github.com/foojank/foojank/clients/repository"
	"github.com/urfave/cli/v2"
	"log/slog"
)

type ListArguments struct {
	Logger     *slog.Logger
	Repository *repository.Client
}

func NewListCommand(args ListArguments) *cli.Command {
	return &cli.Command{
		Name:        "list",
		Description: "List repositories",
		Action:      newListCommandAction(args),
	}
}

func newListCommandAction(args ListArguments) cli.ActionFunc {
	return func(c *cli.Context) error {
		repos, err := args.Repository.List(c.Context)
		if err != nil {
			return err
		}

		for i := range repos {
			fmt.Printf("%#v\n", repos[i])
		}

		return nil
	}
}
