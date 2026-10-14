package create

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/config"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a role",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Name,
				Usage: "set role name",
			},
			&cli.StringFlag{
				Name:  flags.Description,
				Usage: "set role description",
			},
			&cli.StringSliceFlag{
				Name:  flags.Privilege,
				Usage: "set role's privileges",
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
				Name:  flags.Credential,
				Usage: "set credential",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
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

	ctx, err = actions.SetupServer(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, _ *cli.Command) (err error) {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)
	srv := actions.GetServerFromContext(ctx)

	roleName, _ := conf.String(flags.Name)
	roleDesc, _ := conf.String(flags.Description)
	rolePrivs, _ := conf.StringSlice(flags.Privilege)

	client := daemon.New(srv)

	_, err = client.RequestCreateRole(ctx, protodaemon.CreateRoleRequest{
		Name:        roleName,
		Description: roleDesc,
		Privileges:  rolePrivs,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create role: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Role %q has been created!", roleName)

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.Name,
		flags.ServerURL,
		flags.Credential,
	} {
		switch opt {
		case flags.Name:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("role name not configured")
			}
		case flags.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case flags.Credential:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("credential not configured")
			}
		}
	}
	return nil
}
