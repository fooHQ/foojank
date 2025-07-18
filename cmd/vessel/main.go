package main

import (
	"context"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/foohq/foojank/internal/sstls"
	"github.com/foohq/foojank/internal/vessel"
	"github.com/foohq/foojank/internal/vessel/config"
	"github.com/foohq/foojank/internal/vessel/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Debug("started")
	defer log.Debug("stopped")

	usr, err := user.Current()
	if err != nil {
		log.Debug("cannot get computer's username", "error", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Debug("cannot get computer's hostname", "error", err)
	}

	servers := strings.Join(config.Servers, ",")
	nc, err := nats.Connect(
		servers,
		nats.UserJWTAndSeed(config.UserJWT, config.UserKeySeed),
		nats.CustomInboxPrefix("_INBOX_"+config.ServiceName),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ClientTLSConfig(nil, sstls.DecodeCertificateHandler(config.TLSCACertificate)),
		nats.ConnectHandler(func(_ *nats.Conn) {
			log.Debug("connected to the server")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Debug("reconnected to the server")
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Debug("disconnected from the server", "error", err.Error())
		}),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			log.Debug("server error", "error", err.Error())
		}),
	)
	if err != nil {
		log.Debug("cannot connect to the server", "error", err)
		return
	}

	for !nc.IsConnected() {
		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return
		}
	}

	ip, err := nc.GetClientIP()
	if err != nil {
		log.Debug("cannot determine computer's IP address", "error", err)
	}

	err = vessel.New(vessel.Arguments{
		Name:    config.ServiceName,
		Version: config.ServiceVersion,
		Metadata: map[string]string{
			"os":         runtime.GOOS,
			"user":       usr.Username,
			"hostname":   hostname,
			"ip_address": ip.String(),
		},
		Connection: nc,
	}).Start(ctx)
	if err != nil {
		log.Debug("cannot start the agent", "error", err)
		return
	}
}
