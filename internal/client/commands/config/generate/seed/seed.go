package seed

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:   "seed",
		Usage:  "Generate a new seed",
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)
	return createAction(logger)(ctx, c)
}

func createAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		operatorName := fmt.Sprintf("OP%s", nuid.Next())
		accountName := fmt.Sprintf("AC%s", nuid.Next())
		servers := []string{"nats://localhost:4222"}
		seedFile, err := NewOutput(servers, operatorName, accountName)
		if err != nil {
			err = fmt.Errorf("cannot generate a seed file: %v", err)
			logger.Error(err.Error())
			return err
		}

		fmt.Println(seedFile.String())

		return nil
	}
}
