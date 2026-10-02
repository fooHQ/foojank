package activate

import (
	"context"
	"errors"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/authdir"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "activate",
		Usage: "Activate credential",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Token,
				Usage: "set activation token",
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

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	token, _ := conf.String(flags.Token)

	claims, err := jwt.DecodeUserClaims(token)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot decode token: %v", err)
		return err
	}

	pubKey := claims.Subject

	users, err := authdir.ListUsers()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get a list of credentials: %v", err)
		return err
	}

	var activated bool
	for _, name := range users {
		userClaims, err := authdir.GetUserJWT(name)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get credential JWT: %v", err)
			return err
		}

		if userClaims.Subject != pubKey {
			continue
		}

		_, userKey, err := authdir.ReadUser(name)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get credential key: %v", err)
			return err
		}

		err = authdir.UpdateUser(name, token, userKey)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot activate credential: %v", err)
			return err
		}

		activated = true
	}

	if !activated {
		logger.ErrorContext(ctx, "Cannot activate credential: credential not found")
		return errors.New("credential not found")
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.Token,
	} {
		switch opt {
		case flags.Token:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("token not configured")
			}
		}
	}
	return nil
}
