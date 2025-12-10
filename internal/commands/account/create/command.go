package create

import (
	"context"
	"errors"
	"io"
	"os"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/nats-io/jwt/v2"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new account",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Name,
				Usage: "set account name",
			},
		},
		Before:          before,
		Action:          action,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(io.Discard, validateConfiguration)(ctx, c)
	if err != nil {
		ctx, err = actions.LoadFlags(os.Stderr)(ctx, c)
		if err != nil {
			return ctx, err
		}
	}

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) (err error) {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	name, _ := conf.String(flags.Name)
	if name == "" {
		name = petname.Generate(2, "_")
	}

	accountJWT, accountKey, err := auth.NewAccount(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate an account: %v", err)
		return err
	}

	userJWT, userKey, err := auth.NewUser(name, accountKey, jwt.Permissions{})
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate a user: %v", err)
		return err
	}

	_, _, err = auth.ReadAccount(name)
	if !errors.Is(err, auth.ErrAccountNotFound) {
		err = errors.New("account already exists")
		logger.ErrorContext(ctx, "Cannot create account %q: %v", name, err)
		return err
	}

	err = auth.WriteAccount(name, accountJWT, accountKey)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot store account: %v", err)
		return err
	}
	defer func() {
		if err == nil {
			return
		}
		err := auth.DeleteAccount(name)
		if err != nil {
			logger.WarnContext(ctx, "Cannot delete account %q: %v", name, err)
		}
	}()

	err = auth.WriteUser(name, userJWT, userKey)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot store user: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Account %q has been created!", name)

	return nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
