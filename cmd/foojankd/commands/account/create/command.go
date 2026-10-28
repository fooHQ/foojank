package create

import (
	"context"
	"errors"
	"fmt"
	"os"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/flags"
	"github.com/foohq/foojank/internal/authdir"

	"github.com/foohq/foojank/cmd/foojankd/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create account",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Name,
				Usage: "set account name",
			},
			&cli.StringFlag{
				Name:  flags.Description,
				Usage: "set account description",
			},
		},
		Before:          before,
		Action:          action,
		ShellComplete:   actions.CompleteFlags,
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(os.Stderr, validateConfiguration)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.SetupLogger(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.SetupOperator(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, _ *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)
	operator := actions.GetOperatorFromContext(ctx)

	name, _ := conf.String(flags.Name)
	description, _ := conf.String(flags.Description)

	if name == "" {
		name = petname.Generate(2, "_")
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

	accountClaims.Description = description

	accountJWT, err := accountClaims.Encode(operator.KeyPair())
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode account JWT: %v", err)
		return err
	}

	accountKey, err := account.Seed()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode account seed: %v", err)
		return err
	}

	err = authdir.CreateAccount(name, accountJWT, accountKey)
	if err != nil {
		if errors.Is(err, authdir.ErrAccountExists) {
			err = fmt.Errorf("%q already exists", name)
		}
		logger.ErrorContext(ctx, "Cannot create account: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Account %q has been created!", name)

	return nil
}

func validateConfiguration(_ *config.Config) error {
	return nil
}
