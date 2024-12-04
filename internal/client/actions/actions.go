package actions

import (
	"context"
	"fmt"
	"github.com/foohq/foojank/internal/client/flags"
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
	noColor := c.Bool(flags.NoColor)
	return slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:     slog.Level(level),
		NoColor:   noColor,
		AddSource: true,
	}))
}

func NewNATSConnection(ctx context.Context, c *cli.Command, logger *slog.Logger) (*nats.Conn, error) {
	server := c.String(flags.Server)
	userJWT := c.String(flags.UserJWT)
	userNkey := c.String(flags.UserNkey)
	nc, err := nats.Connect(server,
		nats.UserJWTAndSeed(userJWT, userNkey),
		nats.MaxReconnects(-1),
		nats.ConnectHandler(func(nc *nats.Conn) {
			logger.Debug("connected to NATS", "server", server)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("reconnected to NATS", "server", server)
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			logger.Warn("disconnected from NATS", "error", err, "server", server)
		}),
		/*nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
			// TODO: set better error message
			logger.Warn("NATS error ", "error", err, "server", server)
		}),*/
	)
	if err != nil {
		logger.Error("cannot connect to NATS server", "server", server, "error", err)
		return nil, err
	}

	return nc, nil
}
