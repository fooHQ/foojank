package list

import (
	"context"
	"fmt"
	"github.com/foojank/foojank/clients/vessel"
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/urfave/cli/v2"
	"log/slog"
	"time"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "list",
		Description: "List connected agents. The command broadcasts a service discovery message to agents with the specified service name and expects the response to arrive in a given time. Try changing the service name or increasing the default timeout if you are not seeing any connected agents.",
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
	return listAction(logger, client)(c)
}

func listAction(logger *slog.Logger, client *vessel.Client) cli.ActionFunc {
	return func(c *cli.Context) error {
		serviceName := c.String("service-name")
		timeout := c.Duration("timeout")

		ctx, cancel := context.WithTimeout(c.Context, timeout)
		defer cancel()

		outputCh := make(chan vessel.Service)
		go func() {
			select {
			case service := <-outputCh:
				fmt.Printf("%#v\n", service)

			case <-ctx.Done():
				return
			}
		}()

		err := client.Discover(ctx, serviceName, outputCh)
		if err != nil {
			err := fmt.Errorf("discovery request failed: %v", err)
			logger.Error(err.Error())
			return err
		}

		return nil
	}
}
