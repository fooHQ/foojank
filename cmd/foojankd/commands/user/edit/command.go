package edit

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/actions"
	"github.com/foohq/foojank/cmd/foojankd/flags"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/config"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		ArgsUsage: "<name>",
		Usage:     "Edit a user",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Description,
				Usage: "set user description",
			},
			&cli.StringSliceFlag{
				Name:  flags.Privilege,
				Usage: "add a privilege",
			},
			&cli.StringSliceFlag{
				Name:  flags.WithoutPrivilege,
				Usage: "remove a privilege",
			},
			&cli.StringFlag{
				Name:  flags.ServerURL,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:      flags.ServerCertificate,
				Usage:     "set path to server's certificate",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:  flags.Account,
				Usage: "set account",
			},
		},
		Before:          before,
		Action:          action,
		ShellComplete:   actions.CompleteUserName,
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
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)
	srv := actions.GetServerFromContext(ctx)

	description, isDescription := conf.String(flags.Description)
	setPrivileges, _ := conf.StringSlice(flags.Privilege)
	unsetPrivileges, _ := conf.StringSlice(flags.WithoutPrivilege)

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	userName := c.Args().First()

	client := daemon.New(srv)

	_, err := client.RequestUpdateUser(ctx, protodaemon.UpdateUserRequest{
		Name:            userName,
		Description:     description,
		IsDescription:   isDescription,
		SetPrivileges:   setPrivileges,
		UnsetPrivileges: unsetPrivileges,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Cannot update user: %v", err)
		return err
	}

	logger.InfoContext(ctx, "User %q has been updated!", userName)

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.ServerURL,
		flags.Account,
	} {
		switch opt {
		case flags.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case flags.Account:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("account not configured")
			}
		}
	}
	return nil
}
