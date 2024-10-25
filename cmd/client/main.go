package main

import (
	"context"
	"crypto/tls"
	"github.com/foojank/foojank/clients/repository"
	"github.com/foojank/foojank/clients/vessel"
	"github.com/foojank/foojank/internal/application"
	"github.com/foojank/foojank/internal/config"
	tint "github.com/lmittmann/tint"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:   slog.LevelDebug,
		NoColor: false,
	}))

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
		ConnectedCB: func(conn *nats.Conn) {
			logger.Info("connected to NATS", "server", config.NatsURL, "user", config.NatsUser)
		},
		ReconnectedCB: func(conn *nats.Conn) {
			logger.Info("reconnected to NATS", "server", config.NatsURL, "user", config.NatsUser)
		},
		DisconnectedErrCB: func(conn *nats.Conn, err error) {
			logger.Info("disconnected from NATS", "error", err, "server", config.NatsURL, "user", config.NatsUser)
		},
	}

	nc, err := opts.Connect()
	if err != nil {
		logger.Error("cannot connect to NATS server", "error", err, "server", config.NatsURL, "user", config.NatsUser)
		os.Exit(1)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		logger.Error("cannot create a JetStream context", "error", err)
		os.Exit(1)
	}

	err = application.New(
		logger,
		vessel.New(nc),
		repository.New(js),
	).RunContext(ctx, os.Args)
	if err != nil {
		// Error logging is done inside each command no need to have a logger in this place.
		os.Exit(1)
	}
}
