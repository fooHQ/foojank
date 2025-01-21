package master

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
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
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: cannot parse configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
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
		operator, err := auth.NewOperator(operatorName)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		accountName := fmt.Sprintf("AC%s", nuid.Next())
		account, err := auth.NewAccount(accountName, []byte(operator.SigningKey), true)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		systemAccount, err := auth.NewAccount("SYS", []byte(operator.SigningKey), false)
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		output, err := config.NewDefault()
		if err != nil {
			err := fmt.Errorf("cannot generate configuration: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		output.Client.SetAccountJWT(account.JWT)
		output.Client.SetAccountKey(account.Key)

		output.Server.SetOperatorJWT(operator.JWT)
		output.Server.SetOperatorKey(operator.Key)
		output.Server.SetAccountJWT(account.JWT)
		// TODO: account's key is lost here!
		output.Server.SetAccountKey(account.SigningKey)
		output.Server.SetSystemAccountJWT(systemAccount.JWT)
		output.Server.SetSystemAccountKey(systemAccount.Key)
		_, _ = fmt.Fprintln(os.Stdout, output.String())

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.LogLevel == nil {
		return errors.New("log level not configured")
	}

	if conf.NoColor == nil {
		return errors.New("no color not configured")
	}

	return nil
}
