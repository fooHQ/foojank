package list

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/formatter"
	protodaemon "github.com/foohq/foojank/proto/daemon"
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
				Name:  flags.Credential,
				Usage: "set credential",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:          before,
		Action:          action,
		Aliases:         []string{"ls"},
		ShellComplete:   actions.CompleteFlags,
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

	ctx, err = actions.SetupServer(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, _ *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)
	srv := actions.GetServerFromContext(ctx)

	agentName, _ := conf.String(flags.Agent)
	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)

	client := daemon.New(srv)

	agents := make(map[string]protodaemon.Agent)
	var agentID string
	if agentName != "" {
		resp, err := client.GetAgent(ctx, protodaemon.GetAgentRequest{
			AgentRef: agentName,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get agent: %v", err)
			return err
		}

		agentID = resp.Agent.ID
		agents[resp.Agent.ID] = resp.Agent
	} else {
		resp, err := client.ListAgents(ctx, protodaemon.ListAgentsRequest{})
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get a list of agents: %v", err)
			return err
		}

		for _, agent := range resp.Agents {
			agents[agent.ID] = agent
		}
	}

	var jobs []protodaemon.Job
	{
		resp, err := client.ListJobs(ctx, protodaemon.ListJobsRequest{
			AgentRef: agentID,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get a list of jobs: %v", err)
			return err
		}

		jobs = resp.Jobs
	}

	table := formatter.NewTable()
	table.SetHeader([]formatter.Cell{
		formatter.NewStringCell("ID").WithBold(),
		formatter.NewStringCell("AGENT").WithBold(),
		formatter.NewStringCell("COMMAND").WithBold(),
		formatter.NewStringCell("UPDATED AT").WithBold(),
		formatter.NewStringCell("STATUS").WithBold(),
	})

	for _, job := range jobs {
		name := job.AgentID
		if agent, ok := agents[job.AgentID]; ok {
			name = agent.Name
		}

		table.AddRow([]formatter.Cell{
			formatter.NewStringCell(job.ID),
			formatter.NewStringCell(name),
			formatter.NewStringCell(formatArgs(job.Command, job.Args)),
			formatter.NewTimeCell(time.Unix(job.Status.UpdatedAt, 0)).WithFormat("relative"),
			formatter.NewStringCell(strings.ToUpper(job.Status.Status)).WithBold(),
		})
	}

	err := formatter.NewFormatter(
		format,
		formatter.WithNoColor(noColor),
		formatter.WithSortByColumn(3, 1),
	).Write(os.Stdout, table)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot write formatted output: %v", err)
		return err
	}

	return nil
}

func formatArgs(command string, args []string) string {
	return strings.Join(append([]string{command}, args...), " ")
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.ServerURL,
		flags.Credential,
	} {
		switch opt {
		case flags.ServerURL:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("server URL not configured")
			}
		case flags.Credential:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("credential not configured")
			}
		}
	}
	return nil
}
