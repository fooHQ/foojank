package remove

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/foojank/path"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagServer           = flags.Server
	FlagUserJWT          = flags.UserJWT
	FlagUserKey          = flags.UserKey
	FlagTLSCACertificate = flags.TLSCACertificate
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		ArgsUsage: "<storage>:<file>...",
		Usage:     "Remove file from a storage",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  FlagServer,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:  FlagUserJWT,
				Usage: "set user JWT token",
			},
			&cli.StringFlag{
				Name:  FlagUserKey,
				Usage: "set user secret key",
			},
			&cli.StringFlag{
				Name:  FlagTLSCACertificate,
				Usage: "set TLS CA certificate",
			},
		},
		Action:       action,
		Aliases:      []string{"rm"},
		OnUsageError: actions.UsageError,
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	srv, err := server.New(conf.Client.Server, *conf.Client.UserJWT, *conf.Client.UserKey, *conf.Client.TLSCACertificate)
	if err != nil {
		log.Error(ctx, "Cannot connect to the server: %v", err)
		return err
	}

	if c.Args().Len() == 0 {
		log.Error(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	for _, file := range c.Args().Slice() {
		filePath, err := path.Parse(file)
		if err != nil {
			log.Error(ctx, "Invalid path %q: %v.", file, err)
			continue
		}

		if filePath.IsLocal() {
			log.Error(ctx, "Path %q is a local path. Files can only be removed from a storage.", filePath)
			continue
		}

		err = removeFile(ctx, srv, filePath.Storage, filePath.FilePath)
		if err != nil {
			log.Error(ctx, "Cannot delete file %q from a storage %q: %v.", filePath.FilePath, filePath.Storage, err)
			continue
		}
	}

	return nil

}

func removeFile(ctx context.Context, srv *server.Client, name, file string) error {
	storage, err := srv.GetObjectStore(ctx, name)
	if err != nil {
		return fmt.Errorf("cannot open storage: %w", err)
	}
	defer func() {
		_ = storage.Close()
	}()

	err = storage.Wait(ctx)
	if err != nil {
		return fmt.Errorf("cannot synchronize storage: %w", err)
	}

	info, err := storage.Stat(file)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("file %q is a directory", file)
	}

	err = storage.Remove(file)
	if err != nil {
		return err
	}

	return nil
}

func validateConfiguration(conf *config.Config) error {
	if conf.LogLevel == nil {
		return errors.New("log level not configured")
	}

	if conf.NoColor == nil {
		return errors.New("no color not configured")
	}

	if conf.Client == nil {
		return errors.New("client configuration is missing")
	}

	if len(conf.Client.Server) == 0 {
		return errors.New("server not configured")
	}

	if conf.Client.UserJWT == nil {
		return errors.New("user jwt not configured")
	}

	if conf.Client.UserKey == nil {
		return errors.New("user key not configured")
	}

	if conf.Client.TLSCACertificate == nil {
		return errors.New("tls ca certificate not configured")
	}

	return nil
}
