package create

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
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
	agentName, _ := conf.String(flags.Agent)

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

	client := daemon.New(srv)

	command := c.Args().First()
	commandArgs := c.Args().Tail()

	agent, err := client.GetAgent(ctx, agentName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create job: %v", err)
		return err
	}

	job := daemon.JobDirectoryEntry{
		ID:        nuid.Next(),
		AgentID:   agent.ID,
		WorkerID:  nuid.Next(),
		GatewayID: agent.GatewayID,
		Config: daemon.JobConfig{
			Command: command,
			Args:    commandArgs,
		},
		CreatedAt: time.Now().UTC(),
	}

	err = client.PublishStartWorkerRequest(ctx, job)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot publish start worker request: %v", err)
		return err
	}

	err = client.CreateJob(ctx, job)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create job: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Job %q has been created!", job.ID)

	return nil
}

func validateConfiguration(conf *config.Config) error {
	for _, opt := range []string{
		flags.Agent,
		flags.ServerURL,
		flags.Account,
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
		case flags.Account:
			v, ok := conf.String(opt)
			if !ok || v == "" {
				return errors.New("account not configured")
			}
		}
	}
	return nil
}
