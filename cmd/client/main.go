package main

import (
	"context"
	"github.com/foojank/foojank/internal/config"
	"github.com/foojank/foojank/internal/services/client"
	"github.com/nats-io/nats.go"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
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
		log.Fatal(err)
	}

	err = client.New(client.Arguments{
		Connection: nc,
	}).Start(ctx)
	if err != nil {
		panic(err)
	}

}
