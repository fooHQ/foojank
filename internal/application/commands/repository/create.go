package repository

import (
	"fmt"
	"github.com/foojank/foojank/clients/repository"
	"github.com/urfave/cli/v2"
	"log/slog"
)

type CreateArguments struct {
	Logger     *slog.Logger
	Repository *repository.Client
}

func NewCreateCommand(args CreateArguments) *cli.Command {
	return &cli.Command{
		Name:        "create",
		Description: "Create a repository",
		Action:      newCreateCommandAction(args),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "description",
			},
		},
	}
}

func newCreateCommandAction(args CreateArguments) cli.ActionFunc {
	return func(c *cli.Context) error {
		name := c.Args().Get(0)
		description := c.String("description")

		if name == "" {
			return fmt.Errorf("command expects an argument")
		}

		err := args.Repository.Create(c.Context, name, description)
		if err != nil {
			return err
		}

		return nil
	}
}
