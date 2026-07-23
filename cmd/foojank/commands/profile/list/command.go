package list

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List profiles or their details",
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
		Aliases:         []string{"ls"},
		OnUsageError:    actions.UsageError,
		HideHelpCommand: true,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(os.Stderr, validateConfiguration)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.LoadProfiles(os.Stderr)(ctx, c)
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
	profs := actions.GetProfilesFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)

	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)

	profiles := profs.List()

	table := formatter.NewTable()
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell("OS").WithBold(),
		formatter.NewStringCell("ARCH").WithBold(),
	})
	for _, name := range profiles {
		prof, err := profs.Get(name)
		if err != nil {
			return err
		}

		table.AddRow([]formatter.Cell{
			formatter.NewStringCell(name),
			formatter.NewStringCell(prof.OS()),
			formatter.NewStringCell(prof.Arch()),
		})
	}

	err := formatter.NewFormatter(
		format,
		formatter.WithNoColor(noColor),
		formatter.WithSortByColumn(0),
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
