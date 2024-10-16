package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/foojank/foojank/clients/vessel"
	"github.com/foojank/foojank/internal/application"
	"github.com/foojank/foojank/internal/config"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	opts := nats.Options{
		Url:      config.NatsURL,
		User:     config.NatsUser,
		Password: config.NatsPassword,
		// TODO: delete before the release!
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		AllowReconnect: true,
		MaxReconnect:   -1,
	}

	nc, err := opts.Connect()
	if err != nil {
		log.Fatal(err)
	}

	vesselCli := vessel.New(nc)
	app := application.New(vesselCli)
	err = app.RunContext(ctx, os.Args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
