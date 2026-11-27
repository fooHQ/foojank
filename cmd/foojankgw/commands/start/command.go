package start

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/nutsdb/nutsdb"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
	"github.com/foohq/foojank/adapter"
	"github.com/foohq/foojank/clients/gateway"
	"github.com/foohq/foojank/internal/foojankgw"

	"github.com/foohq/foojank/cmd/foojankgw/flags"

	"github.com/foohq/foojank/cmd/foojankgw/actions"

	"github.com/foohq/foojank/internal/config"

	"github.com/foohq/foojank/internal/foojankgw/consumer"
	"github.com/foohq/foojank/internal/foojankgw/handler"
	"github.com/foohq/foojank/internal/foojankgw/publisher"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "start",
		Usage:   "Start gateway",
		Version: foojank.Version(),
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
			&cli.BoolFlag{
				Name:    flags.NoColor,
				Usage:   "disable color output",
				Sources: flags.Env(flags.NoColor),
			},
		},
		Before: before,
		Action: action,
		// TODO: shell completion
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

	ctx, err = actions.SetupServer(os.Stderr)(ctx, c)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) error {
	//conf := actions.GetConfigFromContext(ctx)
	logger := actions.GetLoggerFromContext(ctx)
	srv := actions.GetServerFromContext(ctx)

	gatewayID := srv.UserID()

	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir("/tmp/nutsdb"), // TODO: change me!
	)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot open database: %v", err)
		return err
	}

	// TODO: configure bucket name!
	bucket := "test"
	err = db.Update(func(tx *nutsdb.Tx) error {
		if tx.ExistBucket(nutsdb.DataStructureList, bucket) {
			return nil
		}
		return tx.NewBucket(nutsdb.DataStructureList, bucket)
	})
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create bucket %q: %v", bucket, err)
		return err
	}

	err = foojankgw.New(logger, foojankgw.Config{
		Consumer: consumer.New(logger, consumer.Config{
			Connection: srv.Conn(),
			GatewayID:  gatewayID,
		}),
		Handler: handler.New(logger, handler.Config{
			Gateway: gateway.New(srv),
			DB:      db,
		}),
		Publisher: publisher.New(logger, publisher.Config{
			Connection: srv.Conn(),
		}),
		// TODO: pass adapters
	}).Start(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot start daemon: %v", err)
		return err
	}

	return nil
}

func initPlugins(cmds []string) ([]*plugin.Client, error) {
	clis := make([]*plugin.Client, len(cmds))
	for _, cmd := range cmds {
		clis = append(clis, plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: adapter.HandshakeConfig,
			Plugins: map[string]plugin.Plugin{
				"plugin": &adapter.Plugin{},
			},
			Cmd: exec.Command(cmd),
		}))
	}
	return clis, nil
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
