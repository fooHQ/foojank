package agent

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
		Name:   "agent",
		Usage:  "Generate agent configuration",
		Action: action,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		return err
	}

	logger := actions.NewLogger(ctx, conf)

	return generateAction(logger, conf)(ctx, c)
}

func generateAction(logger *slog.Logger, conf *config.Config) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		username := fmt.Sprintf("AG%s", nuid.Next())
		account := conf.Account
		if account == nil {
			err := fmt.Errorf("cannot generate agent configuration: no account found")
			logger.Error(err.Error())
			return err
		}

		user, err := config.NewUserAgent(username, account.PublicKey, []byte(account.SigningKeySeed))
		if err != nil {
			err := fmt.Errorf("cannot generate agent configuration: %v", err)
			logger.Error(err.Error())
			return err
		}

		output := config.Config{
			Servers: conf.Servers,
			User: &config.Entity{
				JWT:       user.JWT,
				PublicKey: user.PublicKey,
				KeySeed:   user.KeySeed,
			},
		}

		_, _ = fmt.Fprintln(os.Stdout, output.String())

		return nil
	}
}
