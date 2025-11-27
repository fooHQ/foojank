package list

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/formatter"
	jsonformatter "github.com/foohq/foojank/internal/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/formatter/table"
	"github.com/foohq/foojank/internal/profile"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		ArgsUsage: "[name]",
		Usage:     "List profiles or their details",
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
		Before:       before,
		Action:       action,
		Aliases:      []string{"ls"},
		OnUsageError: actions.UsageError,
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

	name := c.Args().First()
	if name != "" {
		err := listProfile(profs, name, format)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get profile %q: %v", name, err)
			return err
		}
		return nil
	}

	err := listProfiles(profs, format)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get a list of profiles: %v", err)
		return err
	}

	return nil
}

func listProfiles(profs *profile.Profiles, format string) error {
	table := formatter.NewTable([]string{
		"name",
		"source_dir",
		"os",
		"arch",
		"features",
	})
	for _, name := range profs.List() {
		prof, err := profs.Get(name)
		if err != nil {
			return err
		}

		table.AddRow([]string{
			name,
			prof.SourceDir(),
			prof.Get(profile.VarOS).Value(),
			prof.Get(profile.VarArch).Value(),
			prof.Get(profile.VarFeatures).Value(),
		})
	}

	return formatOutput(os.Stdout, format, table)
}

func listProfile(profs *profile.Profiles, name, format string) error {
	prof, err := profs.Get(name)
	if err != nil {
		return err
	}

	var env []string
	for k, v := range prof.List() {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	sort.Strings(env)

	row := []string{
		name,
		prof.SourceDir(),
		strings.Join(env, "\n"),
	}

	table := formatter.NewTable([]string{
		"name",
		"source_dir",
		"env",
	})
	table.AddRow(row)

	return formatOutput(os.Stdout, format, table)
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

func validateConfiguration(conf *config.Config) error {
	return nil
}
