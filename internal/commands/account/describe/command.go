package describe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/formatter"
	jsonformatter "github.com/foohq/foojank/internal/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/formatter/table"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "describe",
		ArgsUsage: "<name>",
		Usage:     "Describe account",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
			},
		},
		Before:          before,
		Action:          action,
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

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	name := c.Args().First()

	accountJWT, _, err := auth.ReadAccount(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read account %q: %v", name, err)
		return err
	}

	accountClaims, err := jwt.DecodeAccountClaims(accountJWT)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot decode JWT: %v", err)
		return err
	}

	userJWT, _, err := auth.ReadUser(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read user %q: %v", name, err)
		return err
	}

	userClaims, err := jwt.DecodeUserClaims(userJWT)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot decode JWT: %v", err)
		return err
	}

	accountID := accountClaims.Issuer
	issued := time.Unix(accountClaims.IssuedAt, 0)
	expires := time.Unix(accountClaims.Expires, 0)

	table := formatter.NewTable(nil)
	table.AddRow([]string{"ID", accountID})
	table.AddRow([]string{"Name", name})
	table.AddRow([]string{"Issued at", formatTime(issued)})
	expiresAt := ""
	if isZeroUnixTime(expires) {
		expiresAt = "never"
	} else {
		expiresAt = formatTime(expires)
	}
	table.AddRow([]string{"Expires at", expiresAt})
	if len(accountClaims.SigningKeys) > 0 {
		keys := accountClaims.SigningKeys.Keys()
		table.AddRow([]string{"Dependent accounts", strings.Join(keys, "\n")})
	}
	if userClaims.IssuerAccount != "" {
		table.AddRow([]string{"Linked account", userClaims.IssuerAccount})
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

func isZeroUnixTime(t time.Time) bool {
	return t.Equal(time.Unix(0, 0))
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
