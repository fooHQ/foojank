package list

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/authdir"
	protodaemon "github.com/foohq/foojank/proto/daemon"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List users",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
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
		Aliases:         []string{"ls"},
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

	return ctx, nil
}

func action(ctx context.Context, _ *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	credsName, _ := conf.String(flags.Credential)
	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)

	userJWT, userSeed, err := authdir.ReadUser(credsName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read user %q: %v", credsName, err)
		return err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	client := daemon.New(srv)

	var users []protodaemon.User
	{
		resp, err := client.RequestListUsers(ctx, protodaemon.ListUsersRequest{})
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get a list of users: %v", err)
			return err
		}

		users = resp.Users
	}

	table := formatter.NewTable()
	table.SetHeader([]formatter.Cell{
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("KIND").WithBold(),
		formatter.NewStringCell("DESCRIPTION").WithBold(),
		formatter.NewStringCell("CREATED AT").WithBold(),
	})

	for _, user := range users {
		table.AddRow([]formatter.Cell{
			formatter.NewStringCell(user.Name),
			formatter.NewStringCell(user.Kind),
			formatter.NewStringCell(user.Description),
			formatter.NewTimeCell(time.Unix(user.CreatedAt, 0)),
		})
	}

	err = formatter.NewFormatter(
		format,
		formatter.WithNoColor(noColor),
		formatter.WithSortByColumn(4, 0),
	).Write(os.Stdout, table)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.ServerURL,
		flags.Credential,
	} {
		switch opt {
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
