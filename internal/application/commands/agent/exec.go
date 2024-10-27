package agent

import (
	"bufio"
	"context"
	"fmt"
	"github.com/foojank/foojank/clients/vessel"
	"github.com/foojank/foojank/internal/application/path"
	"github.com/muesli/cancelreader"
	"github.com/urfave/cli/v2"
	"log/slog"
	"os"
	"sync"
)

type ExecArguments struct {
	Logger *slog.Logger
	Vessel *vessel.Client
}

func NewExecCommand(args ExecArguments) *cli.Command {
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
		Action: newExecuteCommandAction(args),
	}
}

func newExecuteCommandAction(args ExecArguments) cli.ActionFunc {
	return func(c *cli.Context) error {
		cnt := c.Args().Len()
		if cnt != 2 {
			err := fmt.Errorf("command '%s' expects the following arguments: %s", c.Command.Name, c.Command.ArgsUsage)
			args.Logger.Error(err.Error())
			return err
		}

		id := c.Args().Get(0)
		serviceName := c.String("service-name")

		pkg := c.Args().Get(1)
		pkgPath, err := path.Parse(pkg)
		if err != nil {
			err := fmt.Errorf("invalid package path '%s': %v", pkg, err)
			args.Logger.Error(err.Error())
			return err
		}

		if pkgPath.IsLocal() {
			err := fmt.Errorf("path '%s' is a local path, executing packages is only possible from a repository", pkgPath)
			args.Logger.Error(err.Error())
			return err
		}

		if pkgPath.IsDir() {
			err := fmt.Errorf("path '%s' is a directory", pkgPath)
			args.Logger.Error(err.Error())
			return err
		}

		ctx := c.Context

		service, err := args.Vessel.GetInfo(ctx, vessel.NewID(serviceName, id))
		if err != nil {
			err := fmt.Errorf("get info request failed: %v", err)
			args.Logger.Error(err.Error())
			return err
		}

		wid, err := args.Vessel.CreateWorker(ctx, service)
		if err != nil {
			err := fmt.Errorf("create worker request failed: %v", err)
			args.Logger.Error(err.Error())
			return err
		}

		defer func() {
			err := args.Vessel.DestroyWorker(context.Background(), service, wid)
			if err != nil {
				err := fmt.Errorf("destroy worker request failed: %v", err)
				args.Logger.Error(err.Error())
			}
		}()

		workerID, err := args.Vessel.GetWorker(ctx, service, wid)
		if err != nil {
			err := fmt.Errorf("get worker request failed: %v", err)
			args.Logger.Error(err.Error())
			return err
		}

		worker, err := args.Vessel.GetInfo(ctx, workerID)
		if err != nil {
			err := fmt.Errorf("get info request failed: %v", err)
			args.Logger.Error(err.Error())
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
			args.Logger.Error(err.Error())
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			code, err := args.Vessel.Execute(ctx, worker, pkgPath.Repository, pkgPath.FilePath, stdinCh, stdoutCh)
			if err != nil {
				err := fmt.Errorf("execute request failed: %v", err)
				args.Logger.Error(err.Error())
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
