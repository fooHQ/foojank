package server

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go"
)

func New(logger *slog.Logger, servers []string, userJWT, userKey string) (*nats.Conn, error) {
	nc, err := nats.Connect(
		strings.Join(servers, ","),
		nats.UserJWTAndSeed(userJWT, userKey),
		nats.MaxReconnects(-1),
		nats.ConnectHandler(func(nc *nats.Conn) {
			logger.Debug("connected to the server")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("reconnected to the server")
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			err = fmt.Errorf("disconnected from the server: %v", err)
			logger.Warn(err.Error())
		}),
		nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
			err = fmt.Errorf("server error: %v", err)
			logger.Warn(err.Error())
		}),
	)
	if err != nil {
		return nil, err
	}

	return nc, nil
}
