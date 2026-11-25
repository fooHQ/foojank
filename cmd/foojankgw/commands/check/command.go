package check

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/daemon"
	protodaemon "github.com/foohq/foojank/proto/daemon"

	"github.com/foohq/foojank/internal/config"

	"github.com/foohq/foojank/cmd/foojankgw/actions"
	"github.com/foohq/foojank/cmd/foojankgw/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "check",
		Usage: "Check gateway health",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flags.ServerURL,
				Usage:   "set server URL",
				Sources: flags.Env(flags.ServerURL),
			},
			&cli.StringFlag{
				Name:      flags.ServerCertificate,
				Usage:     "set path to server's certificate",
				TakesFile: true,
				Sources:   flags.Env(flags.ServerCertificate),
			},
			&cli.StringFlag{
				Name:    flags.Credential,
				Usage:   "set credential",
				Sources: flags.Env(flags.Credential),
			},
		},
		Before: before,
		Action: action,
		// TODO: shell completion!
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
	logger := actions.GetLoggerFromContext(ctx)
	srv := actions.GetServerFromContext(ctx)

	client := daemon.New(srv)

	resp, err := client.CheckHealth(ctx, protodaemon.HealthCheckRequest{})
	if err != nil {
		logger.WarnContext(ctx, "Failure: %v", err)
		return err
	}

	if resp.Status != "ok" {
		logger.WarnContext(ctx, "Failure: status is not OK")
		return errors.New("status not ok")
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
