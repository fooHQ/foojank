package connector

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/sync/errgroup"

	"github.com/foohq/foojank/internal/vessel/connector/consumer"
	"github.com/foohq/foojank/internal/vessel/connector/decoder"
	"github.com/foohq/foojank/internal/vessel/connector/encoder"
	"github.com/foohq/foojank/internal/vessel/connector/publisher"
	"github.com/foohq/foojank/internal/vessel/log"
)

type Arguments struct {
	Servers       []string
	UserJWT       string
	UserKey       string
	Stream        string
	Consumer      string
	Subject       string
	CACertificate string
	OutputCh      chan<- Message
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
	conn, err := connect(
		ctx,
		s.args.Servers,
		s.args.UserJWT,
		s.args.UserKey,
		s.args.CACertificate,
	)
	if err != nil {
		return err
	}

	consumerOutCh := make(chan consumer.Message)
	encoderInCh := make(chan any)
	encoderOutCh := make(chan []byte)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return consumer.New(consumer.Arguments{
			Connection: conn,
			Stream:     s.args.Stream,
			Consumer:   s.args.Consumer,
			ReplyCh:    encoderInCh,
			OutputCh:   consumerOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return decoder.New(decoder.Arguments{
			InputCh:  consumerOutCh,
			OutputCh: s.args.OutputCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return encoder.New(encoder.Arguments{
			InputCh:  encoderInCh,
			OutputCh: encoderOutCh,
		}).Start(groupCtx)
	})
	group.Go(func() error {
		return publisher.New(publisher.Arguments{
			Connection: conn,
			Subject:    s.args.Subject,
			InputCh:    encoderOutCh,
		}).Start(groupCtx)
	})

	return group.Wait()
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
