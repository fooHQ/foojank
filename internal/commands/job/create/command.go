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

	"github.com/foohq/foojank/clients/agent"
	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/internal/actions"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/flags"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		ArgsUsage: "<agent-id> <command> [args]",
		Usage:     "Create a job",
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
				Usage: "set server account",
			},
			&cli.StringFlag{
				Name:  flags.ConfigDir,
				Usage: "set path to a configuration directory",
			},
		},
		Before:       before,
		Action:       action,
		OnUsageError: actions.UsageError,
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

	if c.Args().Len() < 2 {
		logger.ErrorContext(ctx, "Command expects the following arguments: %s", c.ArgsUsage)
		return errors.New("not enough arguments")
	}

	client := agent.New(srv)

	agentID := c.Args().First()
	cmdArgs := c.Args().Tail()

	file := cmdArgs[0]
	var args []string
	if len(cmdArgs) > 1 {
		args = cmdArgs[1:]
	}

	srcPath := file
	storageName := agentID
	dstPath := path.Join("/_cache", nuid.Next())
	err = copyPackage(ctx, srv, srcPath, storageName, dstPath)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot copy package %q to storage %q: %v", srcPath, storageName, err)
		return err
	}

	workerID := nuid.Next()
	file = fmt.Sprintf("nats://%s", dstPath)
	// TODO: env variables
	err = client.StartWorker(ctx, agentID, workerID, file, args, nil)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create job: %v", err)
		return err
	}

	logger.InfoContext(ctx, "Job %q has been created!", workerID)

	return nil
}

func copyPackage(ctx context.Context, srv *server.Client, srcPath, dstStorage, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	storage, err := srv.GetObjectStore(ctx, dstStorage)
	if err != nil {
		return err
	}
	defer func() {
		_ = storage.Close()
	}()

	err = storage.Wait(ctx)
	if err != nil {
		return err
	}

	dstFile, err := storage.Create(dstPath)
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
