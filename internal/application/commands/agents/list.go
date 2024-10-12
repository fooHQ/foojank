package agents

import (
	"context"
	"fmt"
	vesselcli "github.com/foojank/foojank/clients/vessel"
	"github.com/urfave/cli/v2"
	"time"
)

func NewListCommand(vessel *vesselcli.Client) *cli.Command {
	return &cli.Command{
		Name:        "list",
		Description: "List connected agents. The command broadcasts a service discovery message to agents with the specified service name and expects the response to arrive in a given time. Try changing the service name or increasing the default timeout if you are not seeing any connected agents.",
		Action:      newListCommandAction(vessel),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "service-name",
				Value: "vessel",
			},
			&cli.DurationFlag{
				Name:  "timeout",
				Value: 3 * time.Second,
			},
		},
	}
}

func newListCommandAction(vessel *vesselcli.Client) cli.ActionFunc {
	return func(c *cli.Context) error {
		serviceName := c.String("service-name")
		timeout := c.Duration("timeout")

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
