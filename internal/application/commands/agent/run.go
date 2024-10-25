package agent

import (
	"bufio"
	"fmt"
	"github.com/foojank/foojank/clients/vessel"
	"github.com/muesli/cancelreader"
	"github.com/urfave/cli/v2"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type RunArguments struct {
	Logger *slog.Logger
	Vessel *vessel.Client
}

func NewRunCommand(args RunArguments) *cli.Command {
	return &cli.Command{
		Name: "run",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "script",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "service-name",
				Value: "vessel",
			},
		},
		Action: newRunCommandAction(args),
	}
}

func newRunCommandAction(args RunArguments) cli.ActionFunc {
	return func(c *cli.Context) error {
		id := c.String("id")
		script := c.String("script")
		serviceName := c.String("service-name")

		var data []byte
		var err error
		if strings.HasPrefix(script, ".") || strings.HasPrefix(script, "/") {
			data, err = os.ReadFile(script)
		} else {
			panic("reading script from project root is not supported")
		}

		// Check ReadFile error
		if err != nil {
			return err
		}

		ctx := c.Context
		info, err := args.Vessel.GetInfo(ctx, vessel.NewID(serviceName, id))
		if err != nil {
			return err
		}

		wid, err := args.Vessel.CreateWorker(ctx, info)
		if err != nil {
			return err
		}

		workerID, err := args.Vessel.GetWorker(ctx, info, wid)
		if err != nil {
			return err
		}

		worker, err := args.Vessel.GetInfo(ctx, workerID)
		if err != nil {
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
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			code, err := args.Vessel.Execute(ctx, worker, stdinCh, stdoutCh, data)
			if err != nil {
				fmt.Printf("%v\n", err)
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
