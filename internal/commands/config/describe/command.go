package describe

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "describe",
		Usage: "Describe configuration",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:          before,
		Action:          action,
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

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)

	opts := []any{
		&cli.StringFlag{
			Name:  flags.ServerURL,
			Usage: "Server URL",
		},
		&cli.StringFlag{
			Name:  flags.ServerCertificate,
			Usage: "Path to server's certificate",
		},
		&cli.StringFlag{
			Name:  flags.Account,
			Usage: "Account for server authentication",
		},
		&cli.StringFlag{
			Name:  flags.Format,
			Usage: "Output format (ascii, json)",
		},
		&cli.BoolFlag{
			Name:  flags.NoColor,
			Usage: "Color output",
		},
	}

	table := formatter.NewTable()
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("OPTION").WithBold(),
		formatter.NewStringCell("VALUE").WithBold(),
		formatter.NewStringCell("DESCRIPTION").WithBold(),
	})
	for _, opt := range opts {
		var row []formatter.Cell
		switch v := opt.(type) {
		case *cli.StringFlag:
			vv, _ := conf.String(v.Name)
			row = append(row,
				formatter.NewStringCell(v.Name),
				formatter.NewStringCell(vv),
				formatter.NewStringCell(v.Usage),
			)
		case *cli.BoolFlag:
			vv, _ := conf.Bool(v.Name)
			row = append(row,
				formatter.NewStringCell(v.Name),
				formatter.NewBoolCell(vv),
				formatter.NewStringCell(v.Usage),
			)
		}
		table.AddRow(row)
	}

	err := formatter.NewFormatter(format, formatter.WithNoColor(noColor)).Write(os.Stdout, table)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
