package main

import (
	"context"
	"crypto/tls"
	"github.com/foojank/foojank/internal/config"
	"github.com/foojank/foojank/internal/log"
	"github.com/foojank/foojank/internal/services/vessel"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
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
	log.Debug("url=%s user=%s", config.ServerURL, config.ServerUsername)
	defer log.Debug("stopped")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := nats.Options{
		Url:      config.ServerURL,
		User:     config.ServerUsername,
		Password: config.ServerPassword,
		// TODO: delete before the release!
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		AllowReconnect: true,
		MaxReconnect:   -1,
	}

	nc, err := opts.Connect()
	if err != nil {
		log.Debug("cannot connect to NATS: %v", err)
		return
	}

	js, err := jetstream.New(nc)
	if err != nil {
		log.Debug("cannot enable JetStream: %v", err)
		return
	}

	err = vessel.New(vessel.Arguments{
		Name:    config.ServiceName,
		Version: config.ServiceVersion,
		Metadata: map[string]string{
			"os":       runtime.GOOS,
			"user":     usr.Username,
			"hostname": hostname,
		},
		Connection: nc,
		Stream:     js,
	}).Start(ctx)
	if err != nil {
		log.Debug("%v", err)
		return
	}
}
