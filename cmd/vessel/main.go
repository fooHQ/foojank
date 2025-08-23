package main

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

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

func connect(
	ctx context.Context,
	servers []string,
	userJWT,
	userKey,
	caCertificate string,
) (jetstream.JetStream, error) {
	opts := []nats.Option{
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ConnectHandler(func(_ *nats.Conn) {
			log.Debug("connected to the server")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Debug("reconnected to the server")
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				log.Debug("disconnected from the server", "error", err.Error())
			} else {
				log.Debug("disconnected from the server")
			}
		}),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			log.Debug("server error", "error", err.Error())
		}),
	}

	if userJWT != "" && userKey != "" {
		opts = append(opts, nats.UserJWTAndSeed(userJWT, userKey))
	}

	if caCertificate != "" {
		opts = append(opts, nats.ClientTLSConfig(nil, decodeCertificateHandler(caCertificate)))
	}

	nc, err := nats.Connect(strings.Join(servers, ","), opts...)
	if err != nil {
		return nil, err
	}

	for !nc.IsConnected() {
		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return nil, nil
		}
	}

	jetStream, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	return jetStream, nil
}

func decodeCertificateHandler(s string) func() (*x509.CertPool, error) {
	return func() (*x509.CertPool, error) {
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, err
		}

		cert, err := x509.ParseCertificate(b)
		if err != nil {
			return nil, err
		}

		pool := x509.NewCertPool()
		pool.AddCert(cert)
		return pool, nil
	}
}
