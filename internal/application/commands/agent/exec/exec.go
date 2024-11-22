package exec

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/foohq/foojank/clients/vessel"
	"github.com/foohq/foojank/internal/application/actions"
	"github.com/foohq/foojank/internal/application/path"
	"github.com/foohq/foojank/internal/vessel/errcodes"
	"github.com/muesli/cancelreader"
	"github.com/urfave/cli/v3"
	"log/slog"
	"os"
	"sync"
	"time"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "execute",
		ArgsUsage: "<id> <package-path>",
		Usage:     "Run a script on an agent",
		Action:    action,
		Aliases:   []string{"exec"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	logger := actions.NewLogger(ctx, c)

	if c.Args().Len() != 2 {
		err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
		logger.Error(err.Error())
		return err
	}

	nc, err := actions.NewNATSConnection(ctx, c, logger)
	if err != nil {
		return err
	}

	client := vessel.New(nc)
	return execAction(logger, client)(ctx, c)
}

func execAction(logger *slog.Logger, client *vessel.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		arg := c.Args().Get(0)
		id, err := vessel.ParseID(arg)
		if err != nil {
			err := fmt.Errorf("invalid id '%s'", arg)
			logger.Error(err.Error())
			return err
		}

		pkg := c.Args().Get(1)
		pkgPath, err := path.Parse(pkg)
		if err != nil {
			err := fmt.Errorf("invalid package path '%s': %v", pkg, err)
			logger.Error(err.Error())
			return err
		}

		if pkgPath.IsLocal() {
			err := fmt.Errorf("path '%s' is a local path, executing packages is only possible from a repository", pkgPath)
			logger.Error(err.Error())
			return err
		}

		if pkgPath.IsDir() {
			err := fmt.Errorf("path '%s' is a directory", pkgPath)
			logger.Error(err.Error())
			return err
		}

		service, err := client.GetInfo(ctx, id)
		if err != nil {
			err := fmt.Errorf("get info request failed: %v", err)
			logger.Error(err.Error())
			return err
		}

		wid, err := client.CreateWorker(ctx, service)
		if err != nil {
			err := fmt.Errorf("create worker request failed: %v", err)
			logger.Error(err.Error())
			return err
		}

		defer func() {
			err := client.DestroyWorker(context.Background(), service, wid)
			if err != nil {
				err := fmt.Errorf("destroy worker request failed: %v", err)
				logger.Error(err.Error())
			}
		}()

		var attempts = 3
		var workerID vessel.ID
		for attempt := range attempts + 1 {
			var err error
			workerID, err = client.GetWorker(ctx, service, wid)
			if err != nil {
				var errVessel *vessel.Error
				if errors.As(err, &errVessel) && errVessel.Code == errcodes.ErrWorkerStarting && attempt < attempts {
					logger.Debug("get worker request failed", "attempt", attempt+1, "attempts", attempts, "error", err)
					time.Sleep(300 * time.Millisecond)
					continue
				}

				err := fmt.Errorf("get worker request failed: %v", err)
				logger.Error(err.Error())
				return err
			}
		}

		worker, err := client.GetInfo(ctx, workerID)
		if err != nil {
			err := fmt.Errorf("get info request failed: %v", err)
			logger.Error(err.Error())
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
				fmt.Print(string(line))
			}
		}()

		r, err := cancelreader.NewReader(os.Stdin)
		if err != nil {
			err := fmt.Errorf("cannot create a standard input reader %v", err)
			logger.Error(err.Error())
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			code, err := client.Execute(ctx, worker, pkgPath.Repository, pkgPath.FilePath, stdinCh, stdoutCh)
			if err != nil && !errors.Is(err, context.Canceled) {
				err := fmt.Errorf("execute request failed: %v", err)
				logger.Error(err.Error())
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
