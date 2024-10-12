package agents

import (
	"bufio"
	"fmt"
	vesselcli "github.com/foojank/foojank/clients/vessel"
	"github.com/muesli/cancelreader"
	"github.com/urfave/cli/v2"
	"os"
	"sync"
)

func NewRunCommand(vessel *vesselcli.Client) *cli.Command {
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
		Action: newRunCommandAction(vessel),
	}
}

func newRunCommandAction(vessel *vesselcli.Client) cli.ActionFunc {
	return func(c *cli.Context) error {
		id := c.String("id")
		script := c.String("script")
		serviceName := c.String("service-name")

		ctx := c.Context
		info, err := vessel.GetInfo(ctx, vesselcli.NewID(serviceName, id))
		if err != nil {
			return err
		}

		wid, err := vessel.CreateWorker(ctx, info)
		if err != nil {
			return err
		}

		workerID, err := vessel.GetWorker(ctx, info, wid)
		if err != nil {
			return err
		}

		worker, err := vessel.GetInfo(ctx, workerID)
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
			for {
				select {
				case line, ok := <-stdoutCh:
					if !ok {
						return
					}
					fmt.Print(string(line))
				}
			}
		}()

		r, err := cancelreader.NewReader(os.Stdin)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			code, err := vessel.Execute(ctx, worker, stdinCh, stdoutCh, []byte(script))
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
