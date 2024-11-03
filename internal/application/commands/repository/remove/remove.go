package remove

import (
	"context"
	"fmt"
	"github.com/foojank/foojank/clients/repository"
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/foojank/foojank/internal/application/path"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v3"
	"log/slog"
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
	logger := actions.NewLogger(ctx, c)

	if c.Args().Len() == 0 {
		err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
		logger.Error(err.Error())
		return err
	}

	nc, err := actions.NewNATSConnection(ctx, c, logger)
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
