package remove

import (
	"context"
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/foojank/path"
	"github.com/foohq/foojank/internal/log"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		ArgsUsage: "<storage>:<file>...",
		Usage:     "Remove file from a storage",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flags.ServerURL,
				Usage: "set server URL",
			},
			&cli.StringFlag{
				Name:  flags.ServerCertificate,
				Usage: "set server TLS certificate",
			},
			&cli.StringFlag{
				Name:  flags.Account,
				Usage: "set server account",
			},
		},
		Before:       before,
		Action:       action,
		Aliases:      []string{"rm"},
		OnUsageError: actions.UsageError,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(validateConfiguration)(ctx, c)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func action(ctx context.Context, c *cli.Command) error {
	conf := actions.GetConfigFromContext(ctx)

	serverURL, _ := conf.String(flags.ServerURL)
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)

	userJWT, userSeed, err := auth.ReadUser(accountName)
	if err != nil {
		log.Error(ctx, "Cannot read user %q: %v", accountName, err)
		return err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
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
