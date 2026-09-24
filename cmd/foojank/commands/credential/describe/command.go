package describe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/actions"
	"github.com/foohq/foojank/cmd/foojankd/flags"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/authdir"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "describe",
		ArgsUsage: "<name>",
		Usage:     "Describe credential",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
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
	ctx, err := actions.LoadConfig(io.Discard, validateConfiguration)(ctx, c)
	if err != nil {
		return ctx, err
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

	userClaims, err := authdir.GetUserJWT(name)
	if err != nil {
		if errors.Is(err, authdir.ErrUserNotFound) {
			err = fmt.Errorf("%q not found", name)
		}
		logger.ErrorContext(ctx, "Cannot get credential: %v", err)
		return err
	}

	pubKey := userClaims.Issuer
	issued := time.Unix(userClaims.IssuedAt, 0)
	expires := time.Unix(userClaims.Expires, 0)

	table := formatter.NewTable()
	table.SetHeader([]formatter.Cell{
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("PUBLIC KEY").WithBold(),
		formatter.NewStringCell("STATUS").WithBold(),
		formatter.NewStringCell("CREATED AT").WithBold(),
		formatter.NewStringCell("EXPIRES AT").WithBold(),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell(name),
		formatter.NewStringCell(pubKey),
		formatter.NewStringCell(auth.GetJWTStatus(userClaims).String()),
		formatter.NewTimeCell(issued),
		formatter.NewTimeCell(expires).WithEmptyValue("never"),
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

func validateConfiguration(_ *config.Config) error {
	return nil
}
