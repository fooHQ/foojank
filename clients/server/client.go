package server

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/foohq/foojank/internal/sstls"
)

func New(
	logger *slog.Logger,
	servers []string,
	userJWT,
	userKey,
	caCertFile string,
) (*nats.Conn, error) {
	nc, err := nats.Connect(
		strings.Join(servers, ","),
		nats.UserJWTAndSeed(userJWT, userKey),
		nats.MaxReconnects(-1),
		nats.ClientTLSConfig(nil, sstls.DecodeCertificateHandler(caCertFile)),
		nats.ConnectHandler(func(_ *nats.Conn) {
			logger.Debug("connected to the server")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			logger.Info("reconnected to the server")
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			err = fmt.Errorf("disconnected from the server: %w", err)
			logger.Warn(err.Error())
		}),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			err = fmt.Errorf("server error: %w", err)
			logger.Warn(err.Error())
		}),
	)
	if err != nil {
		return nil, err
	}

	return nc, nil
}
