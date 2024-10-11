package commands

import (
	"context"
	"fmt"
	vesselcli "github.com/foojank/foojank/clients/vessel"
	"github.com/urfave/cli/v2"
	"time"
)

func NewListCommand(vessel *vesselcli.Client) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Action: newListCommandAction(vessel),
	}
}

func newListCommandAction(vessel *vesselcli.Client) cli.ActionFunc {
	return func(c *cli.Context) error {
		// TODO: make serviceName configurable!
		serviceName := "vessel"
		// TODO: configurable timeout!
		timeout := 3 * time.Second

		ctx, cancel := context.WithTimeout(c.Context, timeout)
		defer cancel()

		outputCh := make(chan vesselcli.Service)
		go func() {
			select {
			case service := <-outputCh:
				fmt.Printf("%#v\n", service)

			case <-ctx.Done():
				return
			}
		}()

		err := vessel.Discover(ctx, serviceName, outputCh)
		if err != nil {
			return err
		}

		return nil
	}
}
