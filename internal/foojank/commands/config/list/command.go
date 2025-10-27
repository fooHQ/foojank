package list

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/foojank/formatter"
	jsonformatter "github.com/foohq/foojank/internal/foojank/formatter/json"
	tableformatter "github.com/foohq/foojank/internal/foojank/formatter/table"
	"github.com/foohq/foojank/internal/log"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List configuration options",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
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

	format, _ := conf.String(flags.Format)

	opts := []any{
		&cli.StringFlag{
			Name:  flags.ServerURL,
			Usage: "Server URL",
		},
		&cli.StringFlag{
			Name:  flags.ServerCertificate,
			Usage: "Server TLS certificate",
		},
		&cli.StringFlag{
			Name:  flags.Account,
			Usage: "Account for server authentication",
		},
		&cli.StringFlag{
			Name:  flags.Format,
			Usage: "Output format: table or json",
		},
		&cli.BoolFlag{
			Name:  flags.NoColor,
			Usage: "Color output",
		},
	}

	table := formatter.NewTable([]string{
		"option",
		"value",
		"description",
	})
	for _, opt := range opts {
		var name string
		var value string
		var description string

		switch v := opt.(type) {
		case *cli.StringFlag:
			vv, _ := conf.String(v.Name)
			name = v.Name
			value = vv
			description = v.Usage
		case *cli.BoolFlag:
			vv, _ := conf.Bool(v.Name)
			name = v.Name
			value = strconv.FormatBool(vv)
			description = v.Usage
		}

		table.AddRow([]string{
			config.FlagToOption(name),
			value,
			description,
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

func validateConfiguration(conf *config.Config) error {
	return nil
}
