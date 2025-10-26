package generate

import (
	"context"
	"fmt"
	"os"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/nats-io/jwt/v2"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
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

	name, _ := conf.String(flags.Name)
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
	return nil
}
