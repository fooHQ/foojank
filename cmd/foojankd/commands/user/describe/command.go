package describe

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/actions"
	"github.com/foohq/foojank/cmd/foojankd/flags"
	"github.com/foohq/foojank/internal/authdir"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/directory"
	"github.com/foohq/foojank/internal/formatter"
	"github.com/foohq/foojank/internal/handler"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "describe",
		ArgsUsage: "<name>",
		Usage:     "Describe a user",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
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

	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)
	accountName, _ := conf.String(flags.Account)

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	userName := c.Args().First()

	accountKey, err := authdir.GetAccountKey(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read account key: %v", err)
		return err
	}

	userDir, err := directory.OpenUserDirectory(ctx, srv)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot open user directory: %v", err)
		return err
	}

	resp := handler.New(logger, handler.Config{
		UserDirectory: userDir,
		AccountKey:    accountKey,
	}).GetUser(ctx, nil, protodaemon.GetUserRequest{
		UserRef: userName,
	})
	if resp.Error != nil {
		logger.ErrorContext(ctx, "Cannot get user: %v", resp.Error)
		return resp.Error
	}

	user := resp.User

	table := formatter.NewTable()
	table.SetHeader([]formatter.Cell{
		formatter.NewStringCell("ID").WithBold(),
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("DESCRIPTION").WithBold(),
		formatter.NewStringCell("JWT").WithBold(),
		formatter.NewStringCell("PRIVILEGES").WithBold(),
		formatter.NewStringCell("EXPIRES AT").WithBold(),
		formatter.NewStringCell("CREATED AT").WithBold(),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell(user.ID),
		formatter.NewStringCell(user.Name),
		formatter.NewStringCell(user.Description),
		formatter.NewStringCell(user.JWT),
		formatter.NewStringSliceCell(user.Privileges).WithSeparator("\n"),
		formatter.NewTimeCell(time.Unix(user.ExpiresAt, 0)),
		formatter.NewTimeCell(time.Unix(user.CreatedAt, 0)),
	})

	err = formatter.NewFormatter(
		format,
		formatter.WithNoColor(noColor),
		formatter.WithOrientation(formatter.OrientationHorizontal),
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
