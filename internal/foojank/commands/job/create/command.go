package create

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/nats-io/nuid"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/clients/vessel"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
)

const (
	FlagServer           = flags.Server
	FlagUserJWT          = flags.UserJWT
	FlagUserKey          = flags.UserKey
	FlagTLSCACertificate = flags.TLSCACertificate
	FlagDataDir          = flags.DataDir
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		ArgsUsage: "<agent-id> <command> [args]",
		Usage:     "Create a job",
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
			&cli.StringFlag{
				Name:  FlagDataDir,
				Usage: "set path to a data directory",
			},
		},
		Action:       action,
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

	if c.Args().Len() < 2 {
		log.Error(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return err
	}

	client := vessel.New(srv)

	agentID := c.Args().First()
	cmdArgs := c.Args().Tail()

	file := cmdArgs[0]
	var args []string
	if len(cmdArgs) > 1 {
		args = cmdArgs[1:]
	}

	srcPath := file
	repoName := agentID
	dstPath := path.Join("/_cache", nuid.Next())
	err = copyPackage(ctx, srv, srcPath, repoName, dstPath)
	if err != nil {
		log.Error(ctx, "Cannot copy package %q to storage %q: %v", srcPath, repoName, err)
		return err
	}

	workerID := nuid.Next()
	file = fmt.Sprintf("nats://%s", dstPath)
	// TODO: env variables
	err = client.StartWorker(ctx, agentID, workerID, file, args, nil)
	if err != nil {
		log.Error(ctx, "Cannot create job: %v", err)
		return err
	}

	log.Info(ctx, "Job %q has been created!", workerID)

	return nil
}

func copyPackage(ctx context.Context, srv *server.Client, srcPath, dstRepo, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	repo, err := srv.GetObjectStore(ctx, dstRepo)
	if err != nil {
		return err
	}
	defer repo.Close()

	err = repo.Wait(ctx)
	if err != nil {
		return err
	}

	dstFile, err := repo.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

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

	if conf.DataDir == nil {
		return errors.New("data directory not configured")
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

	if conf.DataDir == nil {
		return errors.New("codebase not configured")
	}

	return nil
}
