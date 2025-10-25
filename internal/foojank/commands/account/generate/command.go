package generate

import (
	"context"
	"errors"
	"fmt"
	"os"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/nats-io/jwt/v2"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagName = "name"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate account key and JWT",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  FlagName,
				Usage: "set account name",
			},
		},
		Action:       action,
		OnUsageError: actions.UsageError,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	name := c.Args().First()
	if name == "" {
		name = petname.Generate(2, "_")
	}

	accountJWT, accountKey, err := auth.NewAccount(name)
	if err != nil {
		log.Error(ctx, "Cannot generate an account: %v", err)
		return err
	}

	// TODO: limit user permissions
	// TODO: make atomic write of all changes
	userJWT, userKey, err := auth.NewUser(name, accountKey, jwt.Permissions{})
	if err != nil {
		log.Error(ctx, "Cannot generate a user: %v", err)
		return err
	}

	err = auth.WriteAccount(name, accountJWT, accountKey)
	if err != nil {
		log.Error(ctx, "Cannot store account: %v", err)
		return err
	}

	err = auth.WriteUser(name, userJWT, userKey)
	if err != nil {
		log.Error(ctx, "Cannot store user: %v", err)
		return err
	}

	log.Info(ctx, "Account %q has been created!", name)

	return nil
}

func validateConfiguration(conf *config.Config) error {
	if conf.LogLevel == nil {
		return errors.New("log level not configured")
	}

	if conf.NoColor == nil {
		return errors.New("no color not configured")
	}

	return nil
}
