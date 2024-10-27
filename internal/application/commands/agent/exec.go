package agent

import (
	"bufio"
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

		// TODO: check is directory!

		ctx := c.Context

		info, err := args.Vessel.GetInfo(ctx, vessel.NewID(serviceName, id))
		if err != nil {
			// TODO: create a single error message!
			args.Logger.Error("get info request failed", "error", err)
			return err
		}

		wid, err := args.Vessel.CreateWorker(ctx, info)
		if err != nil {
			args.Logger.Error("create worker request failed", "error", err)
			return err
		}

		workerID, err := args.Vessel.GetWorker(ctx, info, wid)
		if err != nil {
			args.Logger.Error("get worker request failed", "error", err)
			return err
		}

		worker, err := args.Vessel.GetInfo(ctx, workerID)
		if err != nil {
			args.Logger.Error("get info request failed", "error", err)
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
			args.Logger.Error("cannot create a cancel reader", "error", err)
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			code, err := args.Vessel.Execute(ctx, worker, pkgPath.Repository, pkgPath.FilePath, stdinCh, stdoutCh)
			if err != nil {
				args.Logger.Error("execute request failed", "error", err)
				// TODO: handle error!
				//  return error message + code (define which codes should be used!)
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
			case <-ctx.Done():
				return nil
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
