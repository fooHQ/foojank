package edit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		ArgsUsage: "<name>",
		Usage:     "Edit account",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  flags.LinkAccount,
				Usage: "link the specified account to this account and enable shared access to this account's resources",
			},
			&cli.StringFlag{
				Name:  flags.AcceptLinkFrom,
				Usage: "accept a link from the specified account and enable shared access to the specified account’s resources",
			},
			&cli.StringSliceFlag{
				Name:  flags.UnlinkAccount,
				Usage: "remove the link to the specified account and stop shared access to this account’s resources",
			},
			&cli.BoolFlag{
				Name:  flags.RemoveLink,
				Usage: "remove the existing link and stop shared access to the other account’s resources",
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

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	linkAccount, _ := conf.StringSlice(flags.LinkAccount)
	unlinkAccount, _ := conf.StringSlice(flags.UnlinkAccount)
	acceptLinkFrom, _ := conf.String(flags.AcceptLinkFrom)
	removeLink, _ := conf.Bool(flags.RemoveLink)

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	name := c.Args().First()

	{
		accountJWT, accountSeed, err := auth.ReadAccount(name)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot edit account: %v", err)
			return err
		}

		claims, err := jwt.DecodeAccountClaims(accountJWT)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot decode JWT: %v", err)
			return err
		}

		account, err := nkeys.FromSeed(accountSeed)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot decode account seed: %v", err)
			return err
		}

		s := claims.String()
		for _, v := range linkAccount {
			s, err = updateJWTClaims(s, operation{
				action: actionSet,
				path:   "nats.signing_keys.-1",
				value:  v,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Cannot link account %q: %v", v, err)
				return err
			}
		}

		for _, v := range unlinkAccount {
			s, err = updateJWTClaims(s, operation{
				action: actionDelete,
				path:   fmt.Sprintf("nats.signing_keys.#(==%q)", v),
			})
			if err != nil {
				logger.ErrorContext(ctx, "Cannot unlink account %q: %v", v, err)
				return err
			}
		}

		claims = &jwt.AccountClaims{}
		err = json.Unmarshal([]byte(s), &claims)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot decode account claims: %v", err)
			return err
		}

		accountJWT, err = claims.Encode(account)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot encode JWT: %v", err)
			return err
		}

		err = auth.WriteAccount(name, accountJWT, accountSeed)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot edit account: %v", err)
			return err
		}
	}

	{
		userJWT, userSeed, err := auth.ReadUser(name)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot edit account: %v", err)
			return err
		}

		claims, err := jwt.DecodeUserClaims(userJWT)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot decode user claims: %v", err)
			return err
		}

		s := claims.String()
		if acceptLinkFrom != "" {
			s, err = updateJWTClaims(s, operation{
				action: actionSet,
				path:   "nats.issuer_account",
				value:  acceptLinkFrom,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Cannot accept link from %q: %v", acceptLinkFrom, err)
				return err
			}
		}

		if removeLink {
			s, err = updateJWTClaims(s, operation{
				action: actionDelete,
				path:   "nats.issuer_account",
			})
			if err != nil {
				logger.ErrorContext(ctx, "Cannot remove link: %v", err)
				return err
			}
		}

		claims = &jwt.UserClaims{}
		err = json.Unmarshal([]byte(s), &claims)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot decode user claims: %v", err)
			return err
		}

		_, accountSeed, err := auth.ReadAccount(name)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot edit account: %v", err)
			return err
		}

		account, err := nkeys.FromSeed(accountSeed)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot decode account seed: %v", err)
			return err
		}

		userJWT, err = claims.Encode(account)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot encode JWT: %v", err)
			return err
		}

		err = auth.WriteUser(name, userJWT, userSeed)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot edit account: %v", err)
			return err
		}
	}

	return nil
}

const (
	actionSet = iota
	actionDelete
)

type operation struct {
	action int
	path   string
	value  any
}

var errJWTClaimNotFound = errors.New("not found")

func updateJWTClaims(s string, op operation) (string, error) {
	var err error
	res := gjson.Get(s, op.path)
	pth := op.path
	exists := res.Exists()
	if exists {
		pth = res.Path(s)
	}

	switch op.action {
	case actionSet:
		s, err = sjson.Set(s, pth, op.value)
		if err != nil {
			return "", err
		}

	case actionDelete:
		s, err = sjson.Delete(s, pth)
		if err != nil {
			if !exists {
				return "", errJWTClaimNotFound
			}
			return "", err
		}

	default:
		return "", errors.New("unknown operation")
	}

	return s, nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
