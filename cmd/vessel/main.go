package main

import (
	"context"
	"github.com/foojank/foojank/internal/config"
	"github.com/foojank/foojank/internal/log"
	"github.com/foojank/foojank/internal/services/vessel"
	"github.com/foojank/foojank/internal/services/vessel/connector"
	"github.com/foojank/foojank/internal/services/vessel/decoder"
	"github.com/foojank/foojank/internal/services/vessel/scheduler"
	"github.com/nats-io/nats.go"
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

	err = vessel.New(vessel.Arguments{
		Connection: nc,
		Connector: connector.Arguments{
			Name:    config.ConnectorName,
			Version: config.ConnectorVersion,
			Metadata: map[string]string{
				"os":       runtime.GOOS,
				"user":     usr.Username,
				"hostname": hostname,
			},
			Connection: nc,
			OutputCh:   connectorOutCh,
		},
		Decoder: decoder.Arguments{
			InputCh:  connectorOutCh,
			OutputCh: decoderOutCh,
		},
		Scheduler: scheduler.Arguments{
			Connection: nc,
			InputCh:    decoderOutCh,
		},
	}).Start(ctx)
	if err != nil {
		log.Debug("%v", err)
		return
	}
}
