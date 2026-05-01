package list

import (
	"context"
	"errors"
	"os"
	"sort"
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
		Name:  "list",
		Usage: "List jobs",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Agent,
				Usage: "filter jobs by agent",
			},
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
	agentName, _ := conf.String(flags.Agent)
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

	client := agent.New(srv)

	var jobs map[string]agent.Job
	if agentName != "" {
		var agentID string
		agentID, err = client.GetAgentID(ctx, agentName)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get a list of jobs: %v", err)
			return err
		}

		jobs, err = client.ListJobs(ctx, agentID)
	} else {
		jobs, err = client.ListAllJobs(ctx)
	}
	if err != nil {
		logger.ErrorContext(ctx, "Cannot get a list of jobs: %v", err)
		return err
	}

	data := make([]agent.Job, 0, len(jobs))
	for _, job := range jobs {
		data = append(data, job)
	}

	sort.SliceStable(data, func(i, j int) bool {
		return data[i].Updated.Before(data[j].Updated)
	})

	table := formatter.NewTable()
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell("ID").WithBold(),
		formatter.NewStringCell("AGENT").WithBold(),
		formatter.NewStringCell("COMMAND").WithBold(),
		formatter.NewStringCell("LAST UPDATE").WithBold(),
		formatter.NewStringCell("STATUS").WithBold(),
	})
	for _, job := range data {
		table.AddRow([]formatter.Cell{
			formatter.NewStringCell(job.ID),
			formatter.NewStringCell(job.AgentName),
			formatter.NewStringSliceCell([]string{job.Command, job.Args}).WithSeparator(" "),
			formatter.NewTimeCell(job.Updated).WithFormat("relative"),
			formatter.NewStringCell(strings.ToUpper(job.Status)).WithBold(),
		})
	}

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
