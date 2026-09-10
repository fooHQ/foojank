package commands

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/cmd/foojankd/actions"
	"github.com/foohq/foojank/cmd/foojankd/flags"
	"github.com/foohq/foojank/internal/authdir"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/consumer"
	"github.com/foohq/foojank/internal/directory"
	"github.com/foohq/foojank/internal/foojankd"
	"github.com/foohq/foojank/internal/handler"
	"github.com/foohq/foojank/internal/publisher"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "foojankd",
		Usage:   "Daemon for Foojank C2 framework",
		Version: foojank.Version(),
		Flags: []cli.Flag{
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
				Usage: "set account name",
			},
		},
		Before:                before,
		Action:                action,
		CommandNotFound:       actions.CommandNotFound,
		OnUsageError:          actions.UsageError,
		HideHelpCommand:       true,
		EnableShellCompletion: true,
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

	accountKey, err := authdir.GetAccountKey(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read account key: %v", err)
		return err
	}

	userJWT, userKey, err := authdir.ReadUser(accountName)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot read user: %v", err)
		return err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userKey), serverCert)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	agentDir, err := directory.OpenAgentDirectory(ctx, srv)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot open agent directory: %v", err)
		return err
	}

	userDir, err := directory.OpenUserDirectory(ctx, srv)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot open user directory: %v", err)
		return err
	}

	err = foojankd.New(logger, foojankd.Config{
		Consumer: consumer.NewNATSConsumer(logger, consumer.NATSConsumerConfig{
			Connection: srv.Conn(),
		}),
		Handler: handler.NewNATSHandler(logger, handler.NATSHandlerConfig{
			Connection:     srv,
			AgentDirectory: agentDir,
			UserDirectory:  userDir,
			AccountKey:     accountKey,
		}),
		Publisher: publisher.NewNATSPublisher(logger, publisher.NATSPublisherConfig{
			Connection: srv.Conn(),
		}),
	}).Start(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot start daemon: %v", err)
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
