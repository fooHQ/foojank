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
		Name:      "copy",
		ArgsUsage: "[storage:]<file> [storage:]<destination-path>",
		Usage:     "Copy files between local filesystem and a storage or vice versa",
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
		Aliases:      []string{"cp"},
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

	if c.Args().Len() != 2 {
		log.Error(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	files := c.Args().Slice()
	src := files[0]
	srcPath, err := path.Parse(src)
	if err != nil {
		log.Error(ctx, "Invalid path %q: %v.", src, err)
		return err
	}

	dst := files[len(files)-1]
	dstPath, err := path.Parse(dst)
	if err != nil {
		log.Error(ctx, "Invalid path %q: %v.", src, err)
		return err
	}

	if srcPath.IsLocal() && dstPath.IsLocal() {
		log.Error(ctx, "Source and destination paths are both local paths. This operation is currently not supported.")
		return errors.New("matching source and destination type")
	}

	if !srcPath.IsLocal() && !dstPath.IsLocal() {
		log.Error(ctx, "Source and destination paths are both repositories. This operation is currently not supported.")
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
			log.Error(ctx, "Cannot copy file %q to %q: %v", srcPath.String(), dstPath.String(), err)
			return err
		}
		return nil
	}

	// Copy file from a remote storage to a local directory
	if !srcPath.IsLocal() {
		err := copyRemoteFile(ctx, srv, srcPath.Storage, srcPath.FilePath, destPath)
		if err != nil {
			log.Error(ctx, "Cannot copy file %q to %q: %v", srcPath.String(), dstPath.String(), err)
			return err
		}
	}

	return nil
}

func copyLocalFile(ctx context.Context, srv *server.Client, src string, storage, dst string) error {
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

	repo, err := srv.GetObjectStore(ctx, storage)
	if err != nil {
		return fmt.Errorf("cannot open storage: %w", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	err = repo.Wait(ctx)
	if err != nil {
		return fmt.Errorf("cannot synchronize storage: %w", err)
	}

	err = repo.MkdirAll(stdpath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	dstFile, err := repo.Create(dst)
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

func copyRemoteFile(ctx context.Context, srv *server.Client, storage, src string, dst string) error {
	repo, err := srv.GetObjectStore(ctx, storage)
	if err != nil {
		return fmt.Errorf("cannot open storage: %w", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	err = repo.Wait(ctx)
	if err != nil {
		return fmt.Errorf("cannot synchronize storage: %w", err)
	}

	srcFile, err := repo.Open(src)
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
