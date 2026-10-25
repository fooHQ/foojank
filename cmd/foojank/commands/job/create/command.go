package create

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/actions"
	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/config"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		ArgsUsage: "<command> [args]",
		Usage:     "Create a job",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.Agent,
				Usage: "assign the job to the specified agent",
			},
			&cli.StringSliceFlag{
				Name:  flags.Variable,
				Usage: "set environment variable (format: key=value)",
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

func action(ctx context.Context, c *cli.Command) (err error) {
	conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)
	srv := actions.GetServerFromContext(ctx)

	agentName, _ := conf.String(flags.Agent)
	env, _ := conf.StringSlice(flags.Variable)

	if c.Args().Len() < 1 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	client := daemon.New(srv)

	var agent protodaemon.Agent
	{
		resp, err := client.GetAgent(ctx, protodaemon.GetAgentRequest{
			Name: agentName,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Cannot get agent: %v", err)
			return err
		}

		agent = resp.Agent
	}

	resp, err := client.CreateJob(ctx, protodaemon.CreateJobRequest{
		AgentID: agent.ID,
		Command: c.Args().First(),
		Args:    c.Args().Tail(),
		Env:     env,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create job: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Job %q has been created!", resp.JobID)

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.Agent,
		flags.ServerURL,
		flags.Credential,
	} {
		switch opt {
		case flags.Agent:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("agent not configured")
			}
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
