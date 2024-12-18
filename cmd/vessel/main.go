package main

import (
	"context"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/internal/vessel"
	"github.com/foohq/foojank/internal/vessel/config"
	"github.com/foohq/foojank/internal/vessel/log"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Debug("cannot get computer's username: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Debug("cannot get computer's hostname: %v", err)
	}

	log.Debug("started")
	defer log.Debug("stopped")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	servers := strings.Join(config.Servers, ",")
	nc, err := nats.Connect(
		servers,
		nats.UserJWTAndSeed(config.UserJWT, config.UserKeySeed),
		nats.CustomInboxPrefix("_INBOX_"+config.ServiceName),
		nats.MaxReconnects(-1),
		nats.ConnectHandler(func(nc *nats.Conn) {
			log.Debug("connected to the server")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Debug("reconnected to the server")
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			log.Debug(err.Error())
		}),
	)
	if err != nil {
		log.Debug("cannot connect to NATS: %v", err)
		return
	}

	ip, err := nc.GetClientIP()
	if err != nil {
		log.Debug("cannot determine computer's IP address: %v", err)
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
			"os":         runtime.GOOS,
			"user":       usr.Username,
			"hostname":   hostname,
			"ip_address": ip.String(),
		},
		Connection: nc,
		Stream:     js,
	}).Start(ctx)
	if err != nil {
		log.Debug("%v", err)
		return
	}
}
