package list

import (
	"context"
	"fmt"
	"github.com/foojank/foojank/clients/vessel"
	"github.com/foojank/foojank/internal/application/actions"
	"github.com/urfave/cli/v3"
	"log/slog"
	"time"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List active agents",
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
		Action:  action,
		Aliases: []string{"ls"},
	}
}

func action(ctx context.Context, c *cli.Command) error {
	logger := actions.NewLogger(ctx, c)
	nc, err := actions.NewNATSConnection(ctx, c, logger)
	if err != nil {
		return err
	}

	client := vessel.New(nc)
	return listAction(logger, client)(ctx, c)
}

func listAction(logger *slog.Logger, client *vessel.Client) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		serviceName := c.String("service-name")
		timeout := c.Duration("timeout")

		outputCh := make(chan vessel.Service)
		go func() {
			for service := range outputCh {
				fmt.Printf("%#v\n", service)
			}
		}()

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := client.Discover(ctx, serviceName, outputCh)
		if err != nil {
			err := fmt.Errorf("discovery request failed: %v", err)
			logger.Error(err.Error())
			return err
		}

		close(outputCh)

		return nil
	}
}
