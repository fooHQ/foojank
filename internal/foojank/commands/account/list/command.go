package list

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/foojank/formatter"
	jsonformatter "github.com/foohq/foojank/internal/foojank/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/foojank/formatter/table"
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
		Before:       before,
		Action:       action,
		Aliases:      []string{"ls"},
		OnUsageError: actions.UsageError,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(io.Discard, validateConfiguration)(ctx, c)
	if err != nil {
		ctx, err = actions.LoadFlags()(ctx, c)
		if err != nil {
			return ctx, err
		}
	}

	ctx, err = actions.SetupLogger()(ctx, c)
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

	table := formatter.NewTable([]string{
		"name",
		"public_key",
		"created_at",
	})
	for _, account := range accounts {
		accountJWT, _, err := auth.ReadAccount(account)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot read account %q: %v", account, err)
			return err
		}

		claims, err := jwt.DecodeAccountClaims(accountJWT)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot decode JWT: %v", err)
			return err
		}

		ts := time.Unix(claims.IssuedAt, 0)

		table.AddRow([]string{
			claims.Name,
			claims.Subject,
			formatTime(ts),
		})
	}

	err = formatOutput(os.Stdout, format, table)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func formatOutput(w io.Writer, format string, table *formatter.Table) error {
	var f formatter.Formatter
	switch format {
	case "json":
		f = jsonformatter.New()
	case "table":
		f = tableformatter.New()
	default:
		f = tableformatter.New()
	}

	err := f.Write(w, table)
	if err != nil {
		return fmt.Errorf("cannot write formatted output: %w", err)
	}

	return nil
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
