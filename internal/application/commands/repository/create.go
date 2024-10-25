package repository

import (
	"fmt"
	"github.com/foojank/foojank/clients/repository"
	"github.com/urfave/cli/v2"
)

func NewCreateCommand(repo *repository.Client) *cli.Command {
	return &cli.Command{
		Name:        "create",
		Description: "Create a repository",
		Action:      newCreateCommandAction(repo),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "description",
			},
		},
	}
}

func newCreateCommandAction(repo *repository.Client) cli.ActionFunc {
	return func(c *cli.Context) error {
		name := c.Args().Get(0)
		description := c.String("description")

		if name == "" {
			return fmt.Errorf("command expects an argument")
		}

		err := repo.Create(c.Context, name, description)
		if err != nil {
			return err
		}

		return nil
	}
}
