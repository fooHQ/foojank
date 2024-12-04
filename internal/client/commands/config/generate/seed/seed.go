package seed

import (
	"context"
	"fmt"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/urfave/cli/v3"
	"log/slog"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:   "seed",
		Usage:  "Generate a new seed",
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	logger := actions.NewLogger(ctx, c)
	return createAction(logger)(ctx, c)
}

func createAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		// TODO: make operator name configurable - or generate random!
		// TODO: make account name configurable - or generate random!
		// TODO: make servers configurable!
		servers := []string{"nats://localhost:4222"}
		seedFile, err := NewOutput(servers, "operatorTest", "accountTest")
		if err != nil {
			err = fmt.Errorf("cannot generate a seed file: %v", err)
			logger.Error(err.Error())
			return err
		}

		fmt.Println(seedFile.String())

		return nil
	}
}
