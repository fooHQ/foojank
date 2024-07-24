package main

import (
	"context"
	"github.com/foojank/foojank/internal/services/connector"
	"github.com/foojank/foojank/internal/services/runner"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := nats.Options{
		Url:            "SERVER",
		User:           "USER",
		Password:       "PASSWORD",
		AllowReconnect: true,
		MaxReconnect:   -1,
	}

	nc, err := opts.Connect()
	if err != nil {
		// TODO: remove panic
		panic(err)
	}

	connectorOutCh := make(chan connector.Message, 65535)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return connector.New(connector.Arguments{
			Connection: nc,
			OutputCh:   connectorOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return runner.New(runner.Arguments{
			InputCh: connectorOutCh,
		}).Start(groupCtx)
	})

	err = group.Wait()
	if err != nil {
		// TODO: remove panic
		panic(err)
	}
}
