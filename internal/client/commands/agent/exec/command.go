package exec

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/muesli/cancelreader"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/clients/codebase"
	"github.com/foohq/foojank/clients/repository"
	"github.com/foohq/foojank/clients/server"
	"github.com/foohq/foojank/clients/vessel"
	"github.com/foohq/foojank/internal/client/actions"
	"github.com/foohq/foojank/internal/client/flags"
	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/vessel/errcodes"
)

const (
	FlagServer  = "server"
	FlagUserJWT = "user-jwt"
	FlagUserKey = "user-key"
	FlagDataDir = flags.DataDir
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "execute",
		ArgsUsage: "<id> <script-name>",
		Usage:     "Execute a script on an agent",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    FlagServer,
				Usage:   "set server URL",
				Aliases: []string{"s"},
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
				Name:  FlagDataDir,
				Usage: "set path to a data directory",
			},
		},
		Action:  action,
		Aliases: []string{"exec"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	conf, err := actions.NewConfig(ctx, c)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: cannot parse configuration: %v\n", c.FullName(), err)
		return err
	}

	err = validateConfiguration(conf)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid configuration: %v\n", c.FullName(), err)
		return err
	}

	logger := log.New(*conf.LogLevel, *conf.NoColor)

	nc, err := server.New(logger, conf.Client.Server, *conf.Client.UserJWT, *conf.Client.UserKey)
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
	// TODO: this should probably be defined in the config!
	codebaseDir := filepath.Join(*conf.DataDir, "src")
	codebaseCli := codebase.New(codebaseDir)
	repositoryCli := repository.New(js)
	return execAction(logger, vesselCli, codebaseCli, repositoryCli)(ctx, c)
}

func execAction(logger *slog.Logger, vesselCli *vessel.Client, codebaseCli *codebase.Client, repositoryCli *repository.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		if c.Args().Len() < 2 {
			err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		allArgs := c.Args().Slice()
		id := allArgs[0]
		scriptName := allArgs[1]
		// Script arguments should include the name of the script as well.
		scriptArgs := allArgs[1:]

		agentID, err := vessel.ParseID(id)
		if err != nil {
			err := fmt.Errorf("invalid id '%s'", id)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		pkgPath, err := codebaseCli.BuildScript(scriptName)
		if err != nil {
			err := fmt.Errorf("cannot build script '%s': %w", scriptName, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		repoName := agentID.ServiceName()
		err = repositoryCli.Create(ctx, repoName, "")
		if err != nil {
			// FIXME: for some reason existing repository does not cause the method to fail...
			err := fmt.Errorf("cannot create repository '%s': %w", repoName, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		f, err := os.Open(pkgPath)
		if err != nil {
			err := fmt.Errorf("cannot open file '%s': %w", pkgPath, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}
		defer f.Close()

		pkgExecPath := filepath.Join(string(os.PathSeparator), "_cache", filepath.Base(pkgPath))
		err = repositoryCli.PutFile(ctx, repoName, pkgExecPath, f)
		if err != nil {
			err := fmt.Errorf("cannot copy file '%s' to the repository '%s' as '%s': %v", pkgPath, repoName, pkgExecPath, err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := repositoryCli.DeleteFile(ctx, repoName, pkgExecPath)
			if err != nil {
				err := fmt.Errorf("cannot delete file '%s' from the repository '%s': %v", pkgExecPath, repoName, err)
				logger.ErrorContext(ctx, err.Error())
				return
			}
		}()

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

		stdinCh := make(chan []byte, 128)
		stdoutCh := make(chan []byte, 1024)
		exitCh := make(chan int64, 1)

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range stdoutCh {
				_, _ = fmt.Fprint(os.Stdout, string(line))
			}
		}()

		r, err := cancelreader.NewReader(os.Stdin)
		if err != nil {
			err := fmt.Errorf("cannot create a standard input reader %w", err)
			logger.ErrorContext(ctx, err.Error())
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			code, err := vesselCli.Execute(ctx, worker, repoName, pkgExecPath, scriptArgs, stdinCh, stdoutCh)
			if err != nil && !errors.Is(err, context.Canceled) {
				err := fmt.Errorf("execute request failed: %w", err)
				logger.ErrorContext(ctx, err.Error())
			}

			// Cancel stdin scanner to unblock the main loop.
			_ = r.Cancel()
			exitCh <- code
		}()

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text() + "\n"
			select {
			case stdinCh <- []byte(line):
			default:
			}
		}

		cancel()
		close(stdoutCh)
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

	return nil
}
