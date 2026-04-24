package list

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List accounts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
			},
		},
		Before:          before,
		Action:          action,
		Aliases:         []string{"ls"},
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

	format, _ := conf.String(flags.Format)

	accounts, err := auth.ListAccounts()
	if err != nil {
		logger.ErrorContext(ctx, "Cannot list accounts: %v", err)
		return err
	}

	table := formatter.NewTable()
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("ACCOUNT ID").WithBold(),
		formatter.NewStringCell("ISSUED AT").WithBold(),
	})
	for _, account := range accounts {
		claims, err := auth.GetAccountJWT(account)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get account %q JWT: %v", account, err)
			return err
		}

		ts := time.Unix(claims.IssuedAt, 0)
		table.AddRow([]formatter.Cell{
			formatter.NewStringCell(claims.Name),
			formatter.NewStringCell(claims.Subject),
			formatter.NewTimeCell(ts),
		})
	}

	err = formatter.NewFormatter(format).Write(os.Stdout, table)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
