package create

import (
	"context"
	"errors"
	"fmt"
	"os"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/nats-io/jwt/v2"
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
			&cli.Int64Flag{
				Name:  flags.JetStreamMemoryStorage,
				Usage: "set JetStream memory storage limit in bytes (-1 unlimited, 0 disabled)",
				Value: jwt.NoLimit,
			},
			&cli.Int64Flag{
				Name:  flags.JetStreamDiskStorage,
				Usage: "set JetStream disk storage limit in bytes (-1 unlimited, 0 disabled)",
				Value: jwt.NoLimit,
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

	name, ok := conf.String(flags.Name)
	if !ok {
		name = petname.Generate(2, "_")
	}
	description, _ := conf.String(flags.Description)
	memoryStorage, ok := conf.Int64(flags.JetStreamMemoryStorage)
	if !ok {
		memoryStorage = jwt.NoLimit
	}
	diskStorage, _ := conf.Int64(flags.JetStreamDiskStorage)
	if !ok {
		diskStorage = jwt.NoLimit
	}

	account, err := auth.GenerateAccount(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate an account key: %v", err)
		return err
	}

	account.Claims().Description = description
	account.Claims().Limits.MemoryStorage = memoryStorage
	account.Claims().Limits.DiskStorage = diskStorage

	accountJWT, err := account.Claims().Encode(operator.KeyPair())
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode account JWT: %v", err)
		return err
	}

	accountKey, err := account.KeyPair().Seed()
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

func validateConfiguration(conf *config.Config) error {
	return nil
}
