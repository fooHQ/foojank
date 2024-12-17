package client

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "client",
		ArgsUsage: "<config-file>",
		Usage:     "Generate client config from master/client config",
		Action:    action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)

	return generateAction(logger)(ctx, c)
}

func generateAction(logger *slog.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		input, err := config.ParseFile(c.Args().First())
		if err != nil {
			err := fmt.Errorf("cannot parse configuration file: %v", err)
			logger.Error(err.Error())
			return err
		}

		account := input.Account
		if account == nil {
			err := fmt.Errorf("cannot generate client configuration: no account found")
			logger.Error(err.Error())
			return err
		}

		accountClaims, err := jwt.DecodeAccountClaims(account.JWT)
		if err != nil {
			err := fmt.Errorf("cannot build an agent: cannot decode account JWT: %v", err)
			logger.Error(err.Error())
			return err
		}

		accountPubKey := accountClaims.Subject
		username := fmt.Sprintf("MG%s", nuid.Next())
		user, err := config.NewUserManager(username, accountPubKey, []byte(account.SigningKeySeed))
		if err != nil {
			err := fmt.Errorf("cannot generate client configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		output := config.Config{
			Servers: input.Servers,
			Account: &config.Entity{
				JWT:            account.JWT,
				SigningKeySeed: account.SigningKeySeed,
			},
			User: user,
		}

		_, _ = fmt.Fprintln(os.Stdout, output.String())

		return nil
	}
}
