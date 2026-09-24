package list

import (
	"context"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/authdir"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List credentials",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
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
	ctx, err := actions.LoadConfig(io.Discard, validateConfiguration)(ctx, c)
	if err != nil {
		ctx, err = actions.LoadFlags(os.Stderr, validateConfiguration)(ctx, c)
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

func action(ctx context.Context, _ *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)

	users, err := authdir.ListUsers()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot list users: %v", err)
		return err
	}

	table := formatter.NewTable()
	table.SetHeader([]formatter.Cell{
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("PUBLIC KEY").WithBold(),
		formatter.NewStringCell("STATUS").WithBold(),
	})
	for _, user := range users {
		claims, err := authdir.GetUserJWT(user)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get user JWT: %v", err)
			continue
		}

		key, err := authdir.GetUserKey(user)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get user key: %v", err)
			continue
		}

		pubKey, err := key.PublicKey()
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get user public key: %v", err)
			continue
		}

		table.AddRow([]formatter.Cell{
			formatter.NewStringCell(user),
			formatter.NewStringCell(pubKey),
			formatter.NewStringCell(auth.GetJWTStatus(claims).String()),
		})
	}

	err = formatter.NewFormatter(
		format,
		formatter.WithNoColor(noColor),
	).Write(os.Stdout, table)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(_ *config.Config) error {
	return nil
}
