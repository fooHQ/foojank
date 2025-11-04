package copy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	stdpath "path"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/foojank/path"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "copy",
		ArgsUsage: "[storage:]<file> [storage:]<destination-path>",
		Usage:     "Copy files between local filesystem and a storage or vice versa",
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
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:       before,
		Action:       action,
		Aliases:      []string{"cp"},
		OnUsageError: actions.UsageError,
	}
}

func before(ctx context.Context, c *cli.Command) (context.Context, error) {
	ctx, err := actions.LoadConfig(os.Stderr, validateConfiguration)(ctx, c)
	if err != nil {
		return ctx, err
	}

	ctx, err = actions.SetupLogger()(ctx, c)
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

	if c.Args().Len() != 2 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	files := c.Args().Slice()
	src := files[0]
	srcPath, err := path.Parse(src)
	if err != nil {
		logger.ErrorContext(ctx, "Invalid path %q: %v.", src, err)
		return err
	}

	dst := files[len(files)-1]
	dstPath, err := path.Parse(dst)
	if err != nil {
		logger.ErrorContext(ctx, "Invalid path %q: %v.", src, err)
		return err
	}

	if srcPath.IsLocal() && dstPath.IsLocal() {
		logger.ErrorContext(ctx, "Source and destination paths are both local paths. This operation is currently not supported.")
		return errors.New("matching source and destination type")
	}

	if !srcPath.IsLocal() && !dstPath.IsLocal() {
		logger.ErrorContext(ctx, "Source and destination paths are both storages. This operation is currently not supported.")
		return errors.New("matching source and destination type")
	}

	var destPath string
	if dstPath.IsDir() {
		destPath = stdpath.Join("/", dstPath.FilePath, srcPath.Base())
	} else {
		destPath = stdpath.Join("/", dstPath.FilePath)
	}

	// Copy local file to a remote storage
	if srcPath.IsLocal() {
		err := copyLocalFile(ctx, srv, srcPath.FilePath, dstPath.Storage, destPath)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot copy file %q to %q: %v", srcPath.String(), dstPath.String(), err)
			return err
		}
		return nil
	}

	// Copy file from a remote storage to a local directory
	if !srcPath.IsLocal() {
		err := copyRemoteFile(ctx, srv, srcPath.Storage, srcPath.FilePath, destPath)
		if err != nil {
			logger.ErrorContext(ctx, "Cannot copy file %q to %q: %v", srcPath.String(), dstPath.String(), err)
			return err
		}
	}

	return nil
}

func copyLocalFile(ctx context.Context, srv *server.Client, src string, storageName, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return fmt.Errorf("source file is a directory")
	}

	storage, err := srv.GetObjectStore(ctx, storageName)
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

	err = storage.MkdirAll(stdpath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	dstFile, err := storage.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func copyRemoteFile(ctx context.Context, srv *server.Client, storageName, src string, dst string) error {
	storage, err := srv.GetObjectStore(ctx, storageName)
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

	srcFile, err := storage.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return fmt.Errorf("source file is a directory")
	}

	err = os.MkdirAll(stdpath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()

	_, err = io.Copy(dstFile, srcFile)
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
