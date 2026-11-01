package push

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/actions"
	"github.com/foohq/foojank/cmd/foojankd/flags"
	"github.com/foohq/foojank/internal/authdir"
	"github.com/foohq/foojank/internal/clients/resolver"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "push",
		ArgsUsage: "<name>",
		Usage:     "Push account",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flags.ServerURL,
				Usage:   "set server URL",
				Sources: flags.Env(flags.ServerURL),
			},
			&cli.StringFlag{
				Name:      flags.ServerCertificate,
				Usage:     "set path to server's certificate",
				TakesFile: true,
				Sources:   flags.Env(flags.ServerCertificate),
			},
			&cli.StringFlag{
				Name:    flags.SystemAccount,
				Usage:   "set system account",
				Sources: flags.Env(flags.SystemAccount),
			},
		},
		Before:          before,
		Action:          action,
		ShellComplete:   actions.CompleteAccountName,
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

	ctx, err = actions.SetupServer(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) error {
	logger := actions.GetLoggerFromContext(ctx)
	srv := actions.GetServerFromContext(ctx)

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	client := resolver.New(srv)

	name := c.Args().First()

	account, err := authdir.GetAccount(name)
	if err != nil {
		if errors.Is(err, authdir.ErrAccountNotFound) {
			err = fmt.Errorf("%q not found", name)
		}
		logger.ErrorContext(ctx, "Cannot get account: %v", err)
		return err
	}

	err = client.UpdateAccountClaims(ctx, account.JWT())
	if err != nil {
		logger.ErrorContext(ctx, "Cannot push account: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.ServerURL,
		flags.SystemAccount,
	} {
		switch opt {
		case flags.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case flags.SystemAccount:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("system account not configured")
			}
		}
	}
	return nil
}
