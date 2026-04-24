package describe

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
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

	accountClaims, err := auth.GetAccountJWT(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get account JWT: %v", err)
		return err
	}

	userClaims, err := auth.GetUserJWT(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get user JWT: %v", err)
		return err
	}

	accountID := accountClaims.Issuer
	issued := time.Unix(accountClaims.IssuedAt, 0)
	expires := time.Unix(accountClaims.Expires, 0)

	table := formatter.NewTable()
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("ID").WithBold(),
		formatter.NewStringCell(accountID),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell(name),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("ISSUED AT").WithBold(),
		formatter.NewTimeCell(issued),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("EXPIRES AT").WithBold(),
		formatter.NewTimeCell(expires).WithEmptyValue("never"),
	})
	if len(accountClaims.SigningKeys) > 0 {
		keys := accountClaims.SigningKeys.Keys()
		table.AddRow([]formatter.Cell{
			formatter.NewStringCell("DEPENDENT ACCOUNTS").WithBold(),
			formatter.NewStringCell(strings.Join(keys, "\n")),
		})
	}
	if userClaims.IssuerAccount != "" {
		table.AddRow([]formatter.Cell{
			formatter.NewStringCell("LINKED ACCOUNTS").WithBold(),
			formatter.NewStringCell(userClaims.IssuerAccount),
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
