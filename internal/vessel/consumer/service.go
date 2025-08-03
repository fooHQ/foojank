package consumer

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/internal/vessel/log"
)

type Arguments struct {
	Servers           []string
	UserJWT           string
	UserKey           string
	Stream            string
	Consumer          string
	BatchSize         int
	CACertificate     string
	ReconnectInterval time.Duration
	OutputCh          chan<- Message
}

type Service struct {
	args Arguments
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
	}
}

func (s *Service) Start(ctx context.Context) error {
	// This is an initial reconnect interval.
	// A short initial interval guarantees that the service will connect "immediately" after start.
	reconnectInterval := 5 * time.Millisecond

	var jetStream jetstream.JetStream
	var err error

	for {
		select {
		case <-time.After(reconnectInterval):
			// Replace the initial interval with one that the service has been configured with.
			reconnectInterval = s.args.ReconnectInterval
			jetStream, err = connect(s.args.Servers, s.args.UserJWT, s.args.UserKey, s.args.CACertificate)
			if err != nil {
				log.Debug("cannot connect to server", "error", err)
				break
			}

			c, err := jetStream.Consumer(ctx, s.args.Stream, s.args.Consumer)
			if err != nil {
				log.Debug("cannot initialize consumer", "error", err)
				break
			}

			err = fetchMessages(ctx, c, s.args.OutputCh, s.args.BatchSize)
			if err != nil {
				log.Debug("cannot fetch messages", "error", err)
				break
			}

		case <-ctx.Done():
			return nil
		}

		if jetStream != nil {
			jetStream.Conn().Close()
		}
	}
}

func connect(servers []string, userJWT, userKey, caCertificate string) (jetstream.JetStream, error) {
	opts := []nats.Option{
		nats.RetryOnFailedConnect(false),
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

	jetStream, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	return jetStream, nil
}

func fetchMessages(
	ctx context.Context,
	consumer jetstream.Consumer,
	outputCh chan<- Message,
	batchSize int,
) error {
	batch, err := consumer.FetchNoWait(batchSize)
	if err != nil {
		log.Debug("cannot fetch messages", "error", err)
		return err
	}

	for msg := range batch.Messages() {
		if msg == nil {
			break
		}

		select {
		case outputCh <- NewMessage(msg):
		case <-ctx.Done():
			return nil
		}
	}
	return nil
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
