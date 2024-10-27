package exec

import (
	"bufio"
	"context"
	"fmt"
	"github.com/foojank/foojank/clients/vessel"
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/foojank/foojank/internal/application/path"
	"github.com/muesli/cancelreader"
	"github.com/urfave/cli/v2"
	"log/slog"
	"os"
	"sync"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      "execute",
		Args:      true,
		ArgsUsage: "<id> <package-path>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "service-name",
				Value: "vessel",
			},
		},
		Action: action,
	}
}

func action(c *cli.Context) error {
	logger := actions.NewLogger(c)
	nc, err := actions.NewNATSConnection(c, logger)
	if err != nil {
		return err
	}

	client := vessel.New(nc)
	return execAction(logger, client)(c)
}

func execAction(logger *slog.Logger, client *vessel.Client) cli.ActionFunc {
	return func(c *cli.Context) error {
		cnt := c.Args().Len()
		if cnt != 2 {
			err := fmt.Errorf("command '%s' expects the following arguments: %s", c.Command.Name, c.Command.ArgsUsage)
			logger.Error(err.Error())
			return err
		}

		id := c.Args().Get(0)
		serviceName := c.String("service-name")

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

		ctx := c.Context

		service, err := client.GetInfo(ctx, vessel.NewID(serviceName, id))
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

		workerID, err := client.GetWorker(ctx, service, wid)
		if err != nil {
			err := fmt.Errorf("get worker request failed: %v", err)
			logger.Error(err.Error())
			return err
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
			if err != nil {
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

		close(stdoutCh)
		wg.Wait()

		code := <-exitCh
		if code != 0 {
			return cli.Exit("", int(code))
		}

		return nil
	}
}
