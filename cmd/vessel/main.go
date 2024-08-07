package main

import (
	"context"
	"github.com/foojank/foojank/config"
	"github.com/foojank/foojank/internal/log"
	"github.com/foojank/foojank/internal/services/connector"
	"github.com/foojank/foojank/internal/services/decoder"
	"github.com/foojank/foojank/internal/services/scheduler"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"os/user"
	"runtime"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Debug("cannot determine current user")
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Debug("cannot determine hostname")
	}

	log.Debug("started")
	log.Debug("url=%s user=%s", config.NatsURL, config.NatsUser)
	defer log.Debug("stopped")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := nats.Options{
		Url:            config.NatsURL,
		User:           config.NatsUser,
		Password:       config.NatsPassword,
		AllowReconnect: true,
		MaxReconnect:   -1,
	}

	nc, err := opts.Connect()
	if err != nil {
		log.Debug("cannot connect to NATS: %v", err)
		return
	}

	connectorOutCh := make(chan connector.Message, 65535)
	decoderOutCh := make(chan decoder.Message, 65535)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return connector.New(connector.Arguments{
			Name:    config.ConnectorName,
			Version: config.ConnectorVersion,
			Metadata: map[string]string{
				"os":       runtime.GOOS,
				"user":     usr.Username,
				"hostname": hostname,
			},
			Connection: nc,
			OutputCh:   connectorOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return decoder.New(decoder.Arguments{
			InputCh:  connectorOutCh,
			OutputCh: decoderOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return scheduler.New(scheduler.Arguments{
			Connection: nc,
			InputCh:    decoderOutCh,
		}).Start(groupCtx)
	})

	err = group.Wait()
	if err != nil {
		log.Debug("%v", err)
		return
	}
}
