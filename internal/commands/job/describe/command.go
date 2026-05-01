package describe

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/clients/agent"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
	"github.com/foohq/foojank/internal/formatter"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "describe",
		ArgsUsage: "<job-id>",
		Usage:     "Describe job",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Format,
				Usage: "set output format",
			},
			&cli.StringFlag{
				Name:  flags.ServerURL,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:      flags.ServerCertificate,
				Usage:     "set path to server's certificate",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:  flags.Account,
				Usage: "set server account",
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

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)
	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)

	userJWT, userSeed, err := auth.ReadUser(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read user %q: %v", accountName, err)
		return err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	if c.Args().Len() < 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	client := agent.New(srv)

	jobID := c.Args().First()

	job, err := client.GetJob(ctx, jobID)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot describe job: %v", err)
		return err
	}

	table := formatter.NewTable()
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("ID").WithBold(),
		formatter.NewStringCell(job.ID),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("AGENT").WithBold(),
		formatter.NewStringCell(job.AgentName),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("COMMAND").WithBold(),
		formatter.NewStringSliceCell([]string{job.Command, job.Args}).WithSeparator(" "),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("STATUS").WithBold(),
		formatter.NewStringCell(strings.ToUpper(job.Status)),
	})
	if job.Error != nil {
		table.AddRow([]formatter.Cell{
			formatter.NewStringCell("ERROR").WithBold(),
			formatter.NewStringCell(job.Error.Error()),
		})
	}
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("CREATED").WithBold(),
		formatter.NewTimeCell(job.Created),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("UPDATED").WithBold(),
		formatter.NewTimeCell(job.Updated),
	})

	err = formatter.NewFormatter(format, formatter.WithNoColor(noColor)).Write(os.Stdout, table)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.ServerURL,
		flags.Account,
	} {
		switch opt {
		case flags.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case flags.Account:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("account not configured")
			}
		}
	}
	return nil
}
