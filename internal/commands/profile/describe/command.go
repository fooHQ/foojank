package describe

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "describe",
		ArgsUsage: "<name>",
		Usage:     "Describe profile",
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

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	name := c.Args().First()

	prof, err := profs.Get(name)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get profile %q: %v", name, err)
		return err
	}

	envs := make([]string, 0, len(prof.Env()))
	for k, v := range prof.Env() {
		envs = append(envs, fmt.Sprintf("%s = %s", k, v))
	}

	table := formatter.NewTable()
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("NAME").WithBold(),
		formatter.NewStringCell(name),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("OS").WithBold(),
		formatter.NewStringCell(prof.OS()),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("ARCH").WithBold(),
		formatter.NewStringCell(prof.Arch()),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("FEATURES").WithBold(),
		formatter.NewStringCell(strings.Join(prof.Features(), ", ")),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("SOURCE DIR").WithBold(),
		formatter.NewStringCell(prof.SourceDir()),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("ENVIRONMENT").WithBold(),
		formatter.NewStringCell(strings.Join(envs, "\n")),
	})

	err = formatter.NewFormatter(format, formatter.WithNoColor(noColor)).Write(os.Stdout, table)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	return nil
}
