package consumer

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/internal/vessel/log"
)

type Arguments struct {
	Servers       []string
	UserJWT       string
	UserKey       string
	Stream        string
	Consumer      string
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
	jetStream, err := connect(s.args.Servers, s.args.UserJWT, s.args.UserKey, s.args.CACertificate)
	if err != nil {
		log.Debug("cannot connect to server", "error", err)
		return err
	}

	for !jetStream.Conn().IsConnected() {
		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return nil
		}
	}

	consumer, err := jetStream.Consumer(ctx, s.args.Stream, s.args.Consumer)
	if err != nil {
		log.Debug("cannot initialize consumer", "error", err)
		return err
	}

	msgs, err := consumer.Messages()
	if err != nil {
		log.Debug("cannot obtain message context", "error", err)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		forwardMessages(ctx, msgs, s.args.OutputCh)
	}()

	<-ctx.Done()
	msgs.Stop()
	wg.Wait()

	return nil
}

func connect(
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

	jetStream, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	return jetStream, nil
}

func forwardMessages(ctx context.Context, msgs jetstream.MessagesContext, outputCh chan<- Message) {
	for {
		msg, err := msgs.Next()
		if err != nil {
			if errors.Is(err, jetstream.ErrMsgIteratorClosed) {
				return
			}
			continue
		}

		select {
		case outputCh <- NewMessage(msg):
		case <-ctx.Done():
			return
		}
	}
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
