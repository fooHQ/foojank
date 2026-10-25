package describe

import (
	"context"
	"errors"
	"os"
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
		Name:      "describe",
		ArgsUsage: "<job-id>",
		Usage:     "Describe a job",
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
		ShellComplete:   actions.CompleteJobID,
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

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)
	srv := actions.GetServerFromContext(ctx)

	format, _ := conf.String(flags.Format)
	noColor, _ := conf.Bool(flags.NoColor)

	if c.Args().Len() != 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	client := daemon.New(srv)

	jobID := c.Args().First()

	var job protodaemon.Job
	{
		resp, err := client.GetJob(ctx, protodaemon.GetJobRequest{
			JobRef: jobID,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get job: %v", err)
			return err
		}

		job = resp.Job
	}

	var agentName string
	{
		resp, err := client.GetAgent(ctx, protodaemon.GetAgentRequest{
			AgentRef: job.AgentID,
		})
		if err != nil {
			agentName = job.AgentID
		} else {
			agentName = resp.Agent.Name
		}
	}

	var gatewayName string
	{
		resp, err := client.GetGateway(ctx, protodaemon.GetGatewayRequest{
			GatewayRef: job.GatewayID,
		})
		if err != nil {
			gatewayName = job.GatewayID
		} else {
			gatewayName = resp.Gateway.Name
		}
	}

	table := formatter.NewTable()
	table.SetHeader([]formatter.Cell{
		formatter.NewStringCell("ID").WithBold(),
		formatter.NewStringCell("AGENT").WithBold(),
		formatter.NewStringCell("GATEWAY").WithBold(),
		formatter.NewStringCell("WORKER").WithBold(),
		formatter.NewStringCell("COMMAND").WithBold(),
		formatter.NewStringCell("ARGS").WithBold(),
		formatter.NewStringCell("ENV").WithBold(),
		formatter.NewStringCell("STATUS").WithBold(),
		formatter.NewStringCell("ERROR").WithBold(),
		formatter.NewStringCell("CREATED AT").WithBold(),
		formatter.NewStringCell("UPDATED AT").WithBold(),
	})
	table.AddRow([]formatter.Cell{
		formatter.NewStringCell(job.ID),
		formatter.NewStringCell(agentName),
		formatter.NewStringCell(gatewayName),
		formatter.NewStringCell(job.WorkerID),
		formatter.NewStringCell(job.Command),
		formatter.NewStringSliceCell(job.Args).WithSeparator("\n"),
		formatter.NewStringSliceCell(job.Env).WithSeparator("\n"),
		formatter.NewStringCell(job.Status.Status),
		formatter.NewStringCell(job.Status.Error),
		formatter.NewTimeCell(time.Unix(job.CreatedAt, 0)),
		formatter.NewTimeCell(time.Unix(job.Status.UpdatedAt, 0)),
	})

	err := formatter.NewFormatter(
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
