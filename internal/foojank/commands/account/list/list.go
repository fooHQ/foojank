package list

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/formatter"
	jsonformatter "github.com/foohq/foojank/internal/foojank/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/foojank/formatter/table"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagFormat = "format"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List accounts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  FlagFormat,
				Usage: "set output format",
				Value: "table",
			},
		},
		Action:       action,
		Aliases:      []string{"ls"},
		OnUsageError: actions.UsageError,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	format := c.String(FlagFormat)

	accounts, err := auth.ListAccounts()
	if err != nil {
		log.Error(ctx, "Cannot list accounts: %v", err)
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
			log.Error(ctx, "Cannot read account %q: %v", account, err)
			return err
		}

		claims, err := jwt.DecodeAccountClaims(accountJWT)
		if err != nil {
			log.Error(ctx, "Cannot decode JWT: %v", err)
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
		log.Error(ctx, "Cannot write formatted output: %v", err)
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
	if conf.LogLevel == nil {
		return errors.New("log level not configured")
	}

	if conf.NoColor == nil {
		return errors.New("no color not configured")
	}

	return nil
}
