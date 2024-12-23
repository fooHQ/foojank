package remove

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/log"
	"github.com/foohq/foojank/internal/client/path"
	"github.com/foohq/foojank/internal/config"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		ArgsUsage: "<repository>:<file>...",
		Usage:     "Remove file from a repository",
		Action:    action,
		Aliases:   []string{"rm"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c, validateConfiguration)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)

	// TODO: refactor
	nc, err := actions.NewServerConnection(ctx, conf, logger)
	if err != nil {
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		err := fmt.Errorf("cannot create a JetStream context: %v", err)
		logger.Error(err.Error())
		return err
	}

	client := repository.New(js)
	return removeAction(logger, client)(ctx, c)
}

func removeAction(logger *slog.Logger, client *repository.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() == 0 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		for _, file := range c.Args().Slice() {
			filePath, err := path.Parse(file)
			if err != nil {
				err := fmt.Errorf("invalid destination path '%s': %v", file, err)
				logger.Error(err.Error())
				continue
			}

			if filePath.IsLocal() {
				err := fmt.Errorf("path '%s' is a local path, files can only be removed from a repository", filePath)
				logger.Error(err.Error())
				continue
			}

			err = client.DeleteFile(ctx, filePath.Repository, filePath.FilePath)
			if err != nil {
				err := fmt.Errorf("cannot delete file '%s' from a repository '%s': %v", filePath.FilePath, filePath.Repository, err)
				logger.Error(err.Error())
				continue
			}
		}

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.Servers == nil {
		return fmt.Errorf("servers not configured")
	}

	if conf.User == nil {
		return fmt.Errorf("user not configured")
	}

	return nil
}
