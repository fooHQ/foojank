package actions

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/foojank/foojank/internal/application/flags"
	"github.com/lmittmann/tint"
	"github.com/nats-io/nats.go"
	"github.com/urfave/cli/v3"
	"log/slog"
	"os"
)

func CommandNotFound(ctx context.Context, c *cli.Command, s string) {
	logger := NewLogger(ctx, c)
	msg := fmt.Sprintf("command '%s %s' does not exist", c.FullName(), s)
	logger.Error(msg)
	os.Exit(1)
}

func NewLogger(ctx context.Context, c *cli.Command) *slog.Logger {
	level := c.Int(flags.LogLevel)
	return slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:   slog.Level(level),
		NoColor: false,
	}))
}

func NewNATSConnection(ctx context.Context, c *cli.Command, logger *slog.Logger) (*nats.Conn, error) {
	server := c.String(flags.Server)
	user := c.String(flags.Username)
	password := c.String(flags.Password)
	opts := nats.Options{
		Url:      server,
		User:     user,
		Password: password,
		// TODO: delete before the release!
		// TODO: auto-enable if --insecure flag is set!
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		AllowReconnect: true,
		MaxReconnect:   -1,
		ConnectedCB: func(conn *nats.Conn) {
			logger.Info("connected to NATS", "server", server, "username", user)
		},
		ReconnectedCB: func(conn *nats.Conn) {
			logger.Info("reconnected to NATS", "server", server, "username", user)
		},
		DisconnectedErrCB: func(conn *nats.Conn, err error) {
			logger.Warn("disconnected from NATS", "error", err, "server", server, "username", user)
		},
	}

	nc, err := opts.Connect()
	if err != nil {
		logger.Error("cannot connect to NATS server", "error", err, "server", server, "username", user)
		return nil, err
	}

	return nc, nil
}
