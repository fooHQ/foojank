package master

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:   "master",
		Usage:  "Generate new master config",
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

		operator, err := config.NewOperator(operatorName)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		account, err := config.NewAccount(accountName, []byte(operator.SigningKeySeed))
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		systemAccount, err := config.NewAccount("SYS", []byte(operator.SigningKeySeed))
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		output := config.Config{
			Host:          "localhost",
			Servers:       []string{"nats://localhost:4222"},
			Operator:      operator,
			Account:       account,
			SystemAccount: systemAccount,
		}

		_, _ = fmt.Fprintln(os.Stdout, output.String())

		return nil
	}
}
