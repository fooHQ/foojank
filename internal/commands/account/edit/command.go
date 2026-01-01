package edit

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
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
		ArgsUsage: "<account-name>",
		Usage:     "Edit JWT",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  flags.Set,
				Usage: "set option (format: key=value)",
			},
			&cli.StringSliceFlag{
				Name:  flags.Unset,
				Usage: "unset option (format: key)",
			},
			&cli.BoolFlag{
				Name:  flags.User,
				Usage: "edit user JWT",
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

	setOptions, _ := conf.StringSlice(flags.Set)
	unsetOptions, _ := conf.StringSlice(flags.Unset)
	userJWT, _ := conf.Bool(flags.User)

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	name := c.Args().First()

	var err error
	if userJWT {
		err = editUserJWT(name, setOptions, unsetOptions)
	} else {
		err = editAccountJWT(name, setOptions, unsetOptions)
	}
	if err != nil {
		logger.ErrorContext(ctx, "Cannot edit account %q: %v", name, err)
		return err
	}

	return nil
}

func editUserJWT(name string, setOptions, unsetOptions []string) error {
	_, accountSeed, err := auth.ReadAccount(name)
	if err != nil {
		return err
	}

	account, err := nkeys.FromSeed(accountSeed)
	if err != nil {
		return err
	}

	userJWT, userSeed, err := auth.ReadUser(name)
	if err != nil {
		return err
	}

	claims, err := jwt.DecodeUserClaims(userJWT)
	if err != nil {
		return err
	}

	s, err := updateJWTClaims(claims.String(), setOptions, unsetOptions)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(s), &claims)
	if err != nil {
		return err
	}

	userJWT, err = claims.Encode(account)
	if err != nil {
		return err
	}

	err = auth.WriteUser(name, userJWT, userSeed)
	if err != nil {
		return err
	}

	return nil
}

func editAccountJWT(name string, setOptions, unsetOptions []string) error {
	accountJWT, accountSeed, err := auth.ReadAccount(name)
	if err != nil {
		return err
	}

	claims, err := jwt.DecodeAccountClaims(accountJWT)
	if err != nil {
		return err
	}

	account, err := nkeys.FromSeed(accountSeed)
	if err != nil {
		return err
	}

	s, err := updateJWTClaims(claims.String(), setOptions, unsetOptions)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(s), &claims)
	if err != nil {
		return err
	}

	accountJWT, err = claims.Encode(account)
	if err != nil {
		return err
	}

	err = auth.WriteAccount(name, accountJWT, accountSeed)
	if err != nil {
		return err
	}

	return nil
}

func updateJWTClaims(s string, setOptions, unsetOptions []string) (string, error) {
	var err error
	for k, v := range config.ParseKVPairsJSON(setOptions) {
		k = strings.Join([]string{"nats", k}, ".")
		s, err = sjson.Set(s, k, v)
		if err != nil {
			return "", err
		}
	}

	for k := range config.ParseKVPairs(unsetOptions) {
		k = strings.Join([]string{"nats", k}, ".")
		s, err = sjson.Delete(s, k)
		if err != nil {
			return "", err
		}
	}

	return s, nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
