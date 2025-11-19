package generate

import (
	"context"
	"errors"
	"io"
	"os"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/nats-io/jwt/v2"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate account JWT and seed",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Name,
				Usage: "set account name",
			},
		},
		Before:       before,
		Action:       action,
		OnUsageError: actions.UsageError,
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

func action(ctx context.Context, c *cli.Command) error {
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

	// TODO: limit user permissions
	// TODO: make atomic write of all changes
	userJWT, userKey, err := auth.NewUser(name, accountKey, jwt.Permissions{})
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate a user: %v", err)
		return err
	}

	pth, err := auth.AccountPath(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get account path: %v", err)
		return err
	}

	_, err = os.Stat(pth)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.ErrorContext(ctx, "Cannot create account %q: %v", name, err)
		return err
	}
	if err == nil {
		err = errors.New("account already exists")
		logger.ErrorContext(ctx, "Cannot create account %q: %v", name, err)
		return err
	}

	err = auth.WriteAccount(name, accountJWT, accountKey)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot store account: %v", err)
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
