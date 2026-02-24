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

	_, _, err = auth.ReadAccount(name)
	if err == nil {
		err = errors.New("account already exists")
		logger.ErrorContext(ctx, "Cannot create account %q: %v", name, err)
		return err
	}

	account, err := auth.NewAccountKey()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate an account key: %v", err)
		return err
	}

	accountClaims, err := auth.NewAccountJWT(name, account)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate an account JWT: %v", err)
		return err
	}

	user, err := auth.NewUserKey()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate a user key: %v", err)
		return err
	}

	userClaims, err := auth.NewUserJWT(name, jwt.Permissions{}, user)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate a user JWT: %v", err)
		return err
	}

	accountJWT, err := accountClaims.Encode(account)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode account JWT: %v", err)
		return err
	}

	accountKey, err := account.Seed()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode account seed: %v", err)
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

	userJWT, err := userClaims.Encode(account)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode user JWT: %v", err)
		return err
	}

	userKey, err := user.Seed()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode user seed: %v", err)
		return err
	}

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
