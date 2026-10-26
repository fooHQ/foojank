//nolint:staticcheck // preserve the package name
package import_

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/role"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "import",
		Usage: "Import roles from a file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      flags.File,
				Usage:     "set path to an input file",
				TakesFile: true,
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

func action(ctx context.Context, _ *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)
	srv := actions.GetServerFromContext(ctx)

	pth, _ := conf.String(flags.File)

	roles, err := role.ParseFile(pth)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot parse roles in %q: %v", pth, err)
		return err
	}

	client := daemon.New(srv)

	for _, imported := range roles {
		_, err = client.CreateRole(ctx, protodaemon.CreateRoleRequest{
			Name:        imported.Name,
			Description: imported.Description,
			Privileges:  imported.Privileges,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Cannot import role %q: %v", imported.Name, err)
			return err
		}
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.File,
		flags.ServerURL,
		flags.Credential,
	} {
		switch opt {
		case flags.File:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("file not configured")
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
