//nolint:staticcheck // preserve the package name
package import_

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/actions"
	"github.com/foohq/foojank/cmd/foojankd/flags"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/directory"
	"github.com/foohq/foojank/internal/handler"
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
				Name:    flags.Account,
				Usage:   "set account",
				Sources: flags.Env(flags.Account),
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

	roleDir, err := directory.OpenRoleDirectory(ctx, srv)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot open role directory: %v", err)
		return err
	}

	h := handler.New(logger, handler.Config{
		RoleDirectory: roleDir,
	})

	for _, imported := range roles {
		resp := h.CreateRole(ctx, nil, protodaemon.CreateRoleRequest{
			Name:        imported.Name,
			Description: imported.Description,
			Privileges:  imported.Privileges,
		})
		if resp.Error != nil {
			logger.ErrorContext(ctx, "Cannot import role %q: %v", imported.Name, resp.Error)
			return resp.Error
		}
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.File,
		flags.ServerURL,
		flags.Account,
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
		case flags.Account:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("account not configured")
			}
		}
	}
	return nil
}
