package repository

import (
	"fmt"
	"github.com/foojank/foojank/clients/repository"
	"github.com/urfave/cli/v2"
)

func NewListCommand(repo *repository.Client) *cli.Command {
	return &cli.Command{
		Name:        "list",
		Description: "List repositories",
		Action:      newListCommandAction(repo),
	}
}

func newListCommandAction(repo *repository.Client) cli.ActionFunc {
	return func(c *cli.Context) error {
		repos, err := repo.List(c.Context)
		if err != nil {
			return err
		}

		for i := range repos {
			fmt.Printf("%#v\n", repos[i])
		}

		return nil
	}
}
