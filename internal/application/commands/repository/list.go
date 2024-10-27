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
		Args:        true,
		ArgsUsage:   "[repository]",
		Description: "List repositories or their contents.",
		Action:      newListCommandAction(args),
	}
}

func newListCommandAction(args ListArguments) cli.ActionFunc {
	return func(c *cli.Context) error {
		ctx := c.Context

		if c.Args().Len() > 0 {
			for _, r := range c.Args().Slice() {
				files, err := args.Repository.ListFiles(ctx, r)
				if err != nil {
					err := fmt.Errorf("cannot list contents of repository '%s': %v", r, err)
					args.Logger.Error(err.Error())
					return err
				}

				for _, file := range files {
					fmt.Printf("%#v\n", file)
				}
			}
			return nil
		}

		repos, err := args.Repository.List(ctx)
		if err != nil {
			return err
		}

		for i := range repos {
			fmt.Printf("%#v\n", repos[i])
		}

		return nil
	}
}
