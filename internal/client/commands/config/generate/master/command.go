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
	"github.com/foohq/foojank/internal/log"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:   "master",
		Usage:  "Generate new master config",
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c, validateConfiguration)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)
	return createAction(logger)(ctx, c)
}

func createAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		operatorName := fmt.Sprintf("OP%s", nuid.Next())
		operator, err := config.NewOperator(operatorName)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		accountName := fmt.Sprintf("AC%s", nuid.Next())
		account, err := config.NewAccount(accountName, []byte(operator.SigningKeySeed), true)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		systemAccount, err := config.NewAccount("SYS", []byte(operator.SigningKeySeed), false)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		// Using an empty string as a filename to force ParseFile to generate the default configuration.
		output, _ := config.ParseFile("", false)
		output.Operator = operator
		output.Account = account
		output.SystemAccount = systemAccount

		_, _ = fmt.Fprintln(os.Stdout, output.String())

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	// TODO: validate LogLevel and NoColor
	return nil
}
