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
		// TODO: configurable timeout!
		ctx, cancel := context.WithTimeout(c.Context, 2*time.Second)
		defer cancel()

		// TODO: make serviceName configurable!
		serviceName := "vessel"
		result, err := vessel.Discover(ctx, serviceName)
		if err != nil {
			return err
		}

		for i := range result {
			fmt.Printf("%#v\n", result[i])
		}

		return nil
	}
}
