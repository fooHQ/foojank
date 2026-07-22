package describe

import (
	"context"
	"errors"
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
	noColor, _ := conf.Bool(flags.NoColor)

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
	table.SetHeader([]formatter.Cell{
		formatter.NewStringCell("ID").WithBold(),
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("DESCRIPTION").WithBold(),
		formatter.NewStringCell("ISSUED AT").WithBold(),
		formatter.NewStringCell("EXPIRES AT").WithBold(),
		formatter.NewStringCell("LINKED ACCOUNT").WithBold(),
		formatter.NewStringCell("DEPENDENT ACCOUNTS").WithBold(),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell(accountID),
		formatter.NewStringCell(name),
		formatter.NewStringCell(accountClaims.Description),
		formatter.NewTimeCell(issued),
		formatter.NewTimeCell(expires).WithEmptyValue("never"),
		formatter.NewStringCell(userClaims.IssuerAccount),
		formatter.NewStringSliceCell(accountClaims.SigningKeys.Keys()).WithSeparator("\n"),
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
	return nil
}
