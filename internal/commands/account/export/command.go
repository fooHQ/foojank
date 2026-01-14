package export

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "export",
		ArgsUsage: "<name>",
		Usage:     "Export JWT",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  flags.User,
				Usage: "export user JWT",
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

	userJWT, _ := conf.Bool(flags.User)

	if c.Args().Len() < 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	name := c.Args().First()

	var (
		s   string
		err error
	)
	if userJWT {
		s, err = exportUserJWT(name)
	} else {
		s, err = exportAccountJWT(name)
	}
	if err != nil {
		logger.ErrorContext(ctx, "Cannot export JWT: %v", err)
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s\n", s)

	return nil
}

func exportUserJWT(name string) (string, error) {
	userJWT, _, err := auth.ReadUser(name)
	if err != nil {
		return "", err
	}
	return userJWT, nil
}

func exportAccountJWT(name string) (string, error) {
	accountJWT, _, err := auth.ReadAccount(name)
	if err != nil {
		return "", err
	}
	return accountJWT, nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
