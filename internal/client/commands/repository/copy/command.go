package copy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/path"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/log"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "copy",
		ArgsUsage: "[repository:]<file> [repository:]<destination-path>",
		Usage:     "Copy files between local filesystem and a repository or vice versa",
		Action:    action,
		Aliases:   []string{"cp"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c, validateConfiguration)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)

	nc, err := server.New(logger, conf.Servers, conf.User.JWT, conf.User.KeySeed)
	if err != nil {
		err := fmt.Errorf("cannot connect to the server: %w", err)
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		err := fmt.Errorf("cannot create a JetStream context: %w", err)
		logger.ErrorContext(ctx, err.Error())
		return err
	}

	client := repository.New(js)
	return copyAction(logger, client)(ctx, c)
}

func copyAction(logger *slog.Logger, client *repository.Client) cli.ActionFunc {
	// Possible use cases:
	// [Destination is a repository]
	// ./path/to/file repository:/                      => repository:/path/to/file
	// ./path/to/file repository:/test                  => repository:/test
	// ./path/to/file repository:/test/                 => repository:/test/path/to/file
	// ./path/to/file ./path/to/file2 repository:/test  => repository:/test/path/to/file
	//                                                  => repository:/test/path/to/file2
	// ./path/to/file ./path/to/file2 repository:/test/ => repository:/test/path/to/file
	//                                                  => repository:/test/path/to/file2
	// [Destination is a local directory]
	// repository:/path/to/file ./ => file
	// repository:/path/to/file repository:/path/to/file ./ => file (!!! SHOW WARNING THAT THIS WILL OVERWRITE THE FIRST FILE !!!)
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() != 2 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		files := c.Args().Slice()
		src := files[0]
		srcPath, err := path.Parse(src)
		if err != nil {
			err := fmt.Errorf("invalid file path '%s': %w", src, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		dst := files[len(files)-1]
		dstPath, err := path.Parse(dst)
		if err != nil {
			err := fmt.Errorf("invalid destination path '%s': %w", dst, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		if srcPath.IsDir() {
			err := fmt.Errorf("file '%s' is a directory, copying directories is currently not supported", srcPath)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		if srcPath.IsLocal() && dstPath.IsLocal() {
			err := fmt.Errorf("both paths are local paths, this operation is currently not supported")
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		if !srcPath.IsLocal() && !dstPath.IsLocal() {
			err := fmt.Errorf("both paths are repository paths, this operation is currently not supported")
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		// Copy local file to a remote repository
		if srcPath.IsLocal() {
			f, err := os.Open(srcPath.FilePath)
			if err != nil {
				err := fmt.Errorf("cannot open local file: %w", err)
				logger.ErrorContext(ctx, err.Error())
				return err
			}
			defer func() {
				_ = f.Close()
			}()

			var filename string
			if dstPath.IsDir() {
				filename = filepath.Join("/", dstPath.FilePath, srcPath.Base())
			} else {
				filename = filepath.Join("/", dstPath.FilePath)
			}

			logger.Debug("put local file to a repository", "src", srcPath, "repository", dstPath.Repository, "dst", filename)

			err = client.PutFile(ctx, dstPath.Repository, filename, f)
			if err != nil {
				err := fmt.Errorf("cannot put local file '%s' to a repository '%s' as '%s': %v", srcPath, dstPath.Repository, filename, err)
				logger.ErrorContext(ctx, err.Error())
				return err
			}

			return nil
		}

		// TODO
		// Copy file from a remote repository to a local directory
		//if !srcPath.IsLocal() {
		//}

		return nil
	}
}

func validateConfiguration(conf *config.Config) error {
	if conf.Servers == nil {
		return errors.New("servers not configured")
	}

	if conf.User == nil {
		return errors.New("user not configured")
	}

	return nil
}
