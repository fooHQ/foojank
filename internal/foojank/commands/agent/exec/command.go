package exec

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/muesli/cancelreader"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v3"

	renos "github.com/foohq/ren/os"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/clients/vessel"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/foojank/actions"
	"github.com/foohq/foojank/internal/foojank/flags"
	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/vessel/errcodes"
)

const (
	FlagScript           = flags.Script
	FlagServer           = flags.Server
	FlagUserJWT          = flags.UserJWT
	FlagUserKey          = flags.UserKey
	FlagTLSCACertificate = flags.TLSCACertificate
	FlagDataDir          = flags.DataDir
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "execute",
		ArgsUsage: "<id>",
		Usage:     "Execute a script on an agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    FlagScript,
				Usage:   "script to execute",
				Aliases: []string{"s"},
			},
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
		Aliases:      []string{"exec"},
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

	logger := log.New(*conf.LogLevel, *conf.NoColor)

	nc, err := server.New(logger, conf.Client.Server, *conf.Client.UserJWT, *conf.Client.UserKey, *conf.Client.TLSCACertificate)
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

	vesselCli := vessel.New(nc)
	codebaseCli := codebase.New(*conf.DataDir)
	repositoryCli := repository.New(js)
	return execAction(logger, vesselCli, codebaseCli, repositoryCli)(ctx, c)
}

func execAction(logger *slog.Logger, vesselCli *vessel.Client, codebaseCli *codebase.Client, repositoryCli *repository.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		if c.Args().Len() != 1 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		id := c.Args().First()

		// Script arguments should include the name of the script as well.
		var scriptArgs []string
		var scriptName string
		if c.IsSet(FlagScript) {
			scriptArgs = strings.Fields(c.String(FlagScript))
			if len(scriptArgs) != 0 {
				scriptName = scriptArgs[0]
			}
		}

		agentID, err := vessel.ParseID(id)
		if err != nil {
			err := fmt.Errorf("invalid id '%s'", id)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		pkgPath, err := codebaseCli.BuildScript(scriptName)
		if err != nil {
			err := fmt.Errorf("cannot build script: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}
		defer os.Remove(pkgPath)

		b, err := os.ReadFile(pkgPath)
		if err != nil {
			err := fmt.Errorf("cannot read file '%s': %w", pkgPath, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		repoName := agentID.ServiceName()
		repo, err := repositoryCli.Get(ctx, repoName)
		if err != nil {
			err := fmt.Errorf("cannot find repository '%s': %w", repoName, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}
		defer repo.Close()

		err = repo.Wait(ctx)
		if err != nil {
			err := fmt.Errorf("cannot synchronize repository '%s': %w", repoName, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		pkgExecPath := path.Join("/", "_cache", filepath.Base(pkgPath))
		err = repo.WriteFile(pkgExecPath, b, 0644)
		if err != nil {
			err := fmt.Errorf("cannot copy file '%s' to the repository '%s' as '%s': %v", pkgPath, repoName, pkgExecPath, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		service, err := vesselCli.GetInfo(ctx, agentID)
		if err != nil {
			err := fmt.Errorf("get info request failed: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		wid, err := vesselCli.CreateWorker(ctx, service)
		if err != nil {
			err := fmt.Errorf("create worker request failed: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := vesselCli.DestroyWorker(ctx, service, wid)
			if err != nil {
				err := fmt.Errorf("destroy worker request failed: %w", err)
				logger.ErrorContext(ctx, err.Error())
			}
		}()

		var attempts = 3
		var workerID vessel.ID
		for attempt := range attempts + 1 {
			var err error
			workerID, err = vesselCli.GetWorker(ctx, service, wid)
			if err != nil {
				var errVessel *vessel.Error
				if errors.As(err, &errVessel) && errVessel.Code == errcodes.ErrWorkerStarting && attempt < attempts {
					logger.Debug("get worker request failed", "attempt", attempt+1, "attempts", attempts, "error", err)
					time.Sleep(300 * time.Millisecond)
					continue
				}

				err := fmt.Errorf("get worker request failed: %w", err)
				logger.ErrorContext(ctx, err.Error())
				return err
			}
		}

		worker, err := vesselCli.GetInfo(ctx, workerID)
		if err != nil {
			err := fmt.Errorf("get info request failed: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		stdin := renos.NewPipe()
		stdout := renos.NewPipe()
		r, err := cancelreader.NewReader(os.Stdin)
		if err != nil {
			err := fmt.Errorf("cannot create a standard input reader: %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = io.Copy(os.Stdout, stdout)
		}()

		exitCh := make(chan int64, 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			code, err := vesselCli.Execute(ctx, worker, pkgExecPath, scriptArgs, stdin, stdout)
			if err != nil && !errors.Is(err, context.Canceled) {
				err := fmt.Errorf("execute request failed: %w", err)
				logger.ErrorContext(ctx, err.Error())
			}

			// Cancel stdin scanner to unblock the scanner loop.
			_ = r.Cancel()
			exitCh <- code
		}()

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := append(scanner.Bytes(), '\n')
			_, _ = stdin.Write(line)
		}

		wg.Wait()

		code := <-exitCh
		if code != 0 {
			return cli.Exit("", int(code))
		}

		return nil
	}
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

	return nil
}
