package create

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/authdir"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create credential",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Name,
				Usage: "set credential name",
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
	ctx, err := actions.LoadConfig(io.Discard, validateConfiguration)(ctx, c)
	if err != nil {
		ctx, err = actions.LoadFlags(os.Stderr, validateConfiguration)(ctx, c)
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

func action(ctx context.Context, _ *cli.Command) (err error) {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	name, _ := conf.String(flags.Name)

	user, err := auth.NewUserKey()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate a user key: %v", err)
		return err
	}

	account, err := auth.NewAccountKey()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate an account key: %v", err)
		return err
	}

	userClaims, err := auth.NewUserJWT(name, jwt.Permissions{}, user)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot generate a user JWT: %v", err)
		return err
	}

	// Mark the initial user JWT as a dummy. The JWT must be replaced by a JWT signed with a real account key.
	auth.SetDummyUserJWT(userClaims)

	userJWT, err := userClaims.Encode(account)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode user JWT: %v", err)
		return err
	}

	userKey, err := user.Seed()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot encode a user seed: %v", err)
		return err
	}

	err = authdir.WriteUser(name, userJWT, userKey)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot store key: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Credential %q has been created!", name)

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.Name,
	} {
		switch opt {
		case flags.Name:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("key name not configured")
			}
		}
	}
	return nil
}
